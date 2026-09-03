export default {
  async fetch(request, env) {
    const allowed = (env.CORS_ORIGINS || '').split(',').map((s) => s.trim()).filter(Boolean)
    const originHeader = request.headers.get('Origin') || ''
    const allow = allowed.includes(originHeader) ? originHeader : allowed[0] || '*'
    const cors = {
      'Access-Control-Allow-Origin': allow,
      Vary: 'Origin',
    }
    if (request.method === 'OPTIONS') {
      return new Response(null, {
        status: 204,
        headers: {
          ...cors,
          'Access-Control-Allow-Methods': 'GET,OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type,Authorization',
        },
      })
    }
    const json = (data, status = 200) =>
      new Response(JSON.stringify(data), {
        status,
        headers: { ...cors, 'Content-Type': 'application/json' },
      })
    const fail = (message, status) => json({ error: message, message }, status)
    if (request.method !== 'GET') return fail('method not allowed', 405)

    const url = new URL(request.url)
    const path = url.pathname.replace(/\/$/, '') || '/'
    const client =
      request.headers.get('CF-Connecting-IP') ||
      (request.headers.get('X-Forwarded-For') || '').split(',')[0].trim() ||
      '0.0.0.0'

    if (path !== '/api/v1/health' && path !== '/api/v1/docs') {
      try {
        await env.DB.prepare('INSERT INTO api_request (client_ip) VALUES (?1)').bind(client).run()
      } catch (e) {}
    }

    if (path === '/api/v1/health') return json({ status: 'ok', service: 'as-ip' })
    if (path === '/api/v1/docs') return json({ service: 'as-ip', base: '/api/v1' })
    if (path === '/api/v1/http/headers') {
      const headers = []
      for (const [name, value] of request.headers.entries()) {
        const lower = name.toLowerCase()
        headers.push({
          name,
          value: lower === 'cookie' || lower === 'authorization' ? '[redacted]' : value,
        })
      }
      return json({ ip: client, headers })
    }

    const isV4 = (ip) => {
      const p = ip.split('.')
      return p.length === 4 && p.every((n) => {
        const v = Number(n)
        return Number.isInteger(v) && v >= 0 && v <= 255
      })
    }

    if (path === '/api/v1/ip/info' || path.startsWith('/api/v1/ip/info/')) {
      const ip = path === '/api/v1/ip/info' ? client : decodeURIComponent(path.slice(16))
      if (!ip.trim()) return fail('ip is required', 400)
      if (!isV4(ip)) return fail('no ASN mapping found for IPv4 address ' + ip, 404)
      return fail('no ASN mapping found for IPv4 address ' + ip, 404)
    }

    if (path === '/api/v1/ip/whois' || path.startsWith('/api/v1/ip/whois/')) {
      const ip = path === '/api/v1/ip/whois' ? client : decodeURIComponent(path.slice(17))
      if (!ip.trim()) return fail('ip is required', 400)
      const rdapUrl = 'https://rdap.org/ip/' + encodeURIComponent(ip)
      const resp = await fetch(rdapUrl, { headers: { Accept: 'application/rdap+json, application/json' } })
      if (resp.status === 404) return fail('no whois/rdap record found for ' + ip, 404)
      if (!resp.ok) return fail('rdap upstream returned HTTP ' + resp.status, 502)
      const raw = await resp.json()
      const entities = []
      const walk = (list) => {
        for (const item of list || []) {
          if (item.handle) entities.push(item.handle)
          if (Array.isArray(item.entities)) walk(item.entities)
        }
      }
      walk(raw.entities)
      return json({
        ip,
        handle: raw.handle || '',
        name: raw.name || '',
        type: raw.type || '',
        country: raw.country || '',
        startAddress: raw.startAddress || '',
        endAddress: raw.endAddress || '',
        status: raw.status || [],
        entities,
        events: (raw.events || []).map((e) => ({ action: e.eventAction || '', date: e.eventDate || '' })),
        remarks: [],
        rdapUrl,
      })
    }

    if (path.startsWith('/api/v1/dns/lookup/')) {
      const domain = decodeURIComponent(path.slice(19)).replace(/^https?:\/\//, '').split('/')[0].replace(/\.$/, '').toLowerCase()
      if (!domain) return fail('domain is required', 400)
      const doh = async (type) => {
        const resp = await fetch('https://cloudflare-dns.com/dns-query?name=' + encodeURIComponent(domain) + '&type=' + type, {
          headers: { Accept: 'application/dns-json' },
        })
        if (!resp.ok) return []
        const data = await resp.json()
        const map = { A: 1, NS: 2, CNAME: 5, MX: 15, TXT: 16, AAAA: 28 }
        const wanted = map[type]
        return (data.Answer || []).filter((a) => a.type === wanted).map((a) => String(a.data).replace(/^"|"$/g, '').replace(/\.$/, ''))
      }
      const [a, aaaa, ns, mx, txt, cname] = await Promise.all([doh('A'), doh('AAAA'), doh('NS'), doh('MX'), doh('TXT'), doh('CNAME')])
      if (![a, aaaa, ns, mx, txt, cname].some((x) => x.length)) return fail('no DNS records found for ' + domain, 404)
      return json({
        domain,
        a,
        aaaa,
        ns,
        mx: mx.map((v) => {
          const parts = v.split(/\s+/, 2)
          return { host: parts[1] || v, pref: parseInt(parts[0], 10) || 0 }
        }),
        txt,
        cname: cname[0] || '',
        asn: 0,
        as: '',
        country: '',
        addresses: a.map((ip) => ({ ip, asn: 0, as: '', country: '' })),
      })
    }

    if (path === '/api/v1/asn/list') {
      const { results } = await env.DB.prepare('SELECT asn AS number, "as" AS name FROM asn ORDER BY asn').all()
      return json((results || []).map((r) => ({ asn: r.number, as: r.name })))
    }
    if (path.startsWith('/api/v1/asn/search/')) return fail('ASN not found', 404)
    if (path.startsWith('/api/v1/asn/to-as/')) return fail('ASN not found', 404)
    if (path.startsWith('/api/v1/as/search/')) return fail('AS not found', 404)
    if (path === '/api/v1/country/list') {
      const { results } = await env.DB.prepare('SELECT country_code AS code, country AS name FROM country ORDER BY country_code').all()
      return json((results || []).map((r) => ({ code: r.code, name: r.name })))
    }
    if (path.startsWith('/api/v1/country/search/')) {
      const code = decodeURIComponent(path.slice(23)).trim()
      if (!/^[a-zA-Z]{2}$/.test(code)) return fail('country must be an ISO 3166-1 alpha-2 code', 400)
      return json([])
    }
    return fail('not found', 404)
  },
}
