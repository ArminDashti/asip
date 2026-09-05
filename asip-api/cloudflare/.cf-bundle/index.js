var __defProp = Object.defineProperty;
var __name = (target, value) => __defProp(target, "name", { value, configurable: true });

// src/index.ts
var AS_SELECT = `SELECT a.id, a."as" AS as_name, cat.category, a.registered, orig.origin, a.last_modified, a.last_announced, c.country_code, c.country, a.asn, COALESCE(s.prefixes, 0) AS prefixes, COALESCE(s.prefixes_aggregated, 0) AS prefixes_aggregated, s.largest_prefix, COALESCE(s.total_addresses, 0) AS total_addresses, COALESCE((SELECT GROUP_CONCAT(provider.asn) FROM connectivity conn JOIN asn provider ON provider.id = conn.provider_asn_id WHERE conn.asn_id = a.id), '') AS provider_asns FROM asn a LEFT JOIN country c ON c.id = a.country_id LEFT JOIN category cat ON cat.id = a.category_id LEFT JOIN origin orig ON orig.id = a.origin_id LEFT JOIN ipv4_stat s ON s.asn_id = a.id`;
function corsOrigin(env, origin) {
  const allowed = (env.CORS_ORIGINS || "").split(",").map((o) => o.trim()).filter(Boolean);
  return allowed.includes(origin) ? origin : allowed[0] || "*";
}
__name(corsOrigin, "corsOrigin");
function withCors(res, origin) {
  const headers = new Headers(res.headers);
  headers.set("Access-Control-Allow-Origin", origin);
  headers.set("Vary", "Origin");
  return new Response(res.body, { status: res.status, headers });
}
__name(withCors, "withCors");
function json(data, status = 200) {
  return Response.json(data, { status });
}
__name(json, "json");
function err(message, status) {
  return json({ error: message, message }, status);
}
__name(err, "err");
function clientIp(req) {
  return req.headers.get("CF-Connecting-IP") || req.headers.get("X-Forwarded-For")?.split(",")[0]?.trim() || "0.0.0.0";
}
__name(clientIp, "clientIp");
function isIpv4(ip) {
  const p = ip.split(".");
  return p.length === 4 && p.every((x) => {
    const n = Number(x);
    return Number.isInteger(n) && n >= 0 && n <= 255 && String(n) === String(Number(x));
  });
}
__name(isIpv4, "isIpv4");
function ipv4ToInt(ip) {
  if (!isIpv4(ip)) return null;
  return ip.split(".").reduce((acc, part) => (acc << 8) + Number(part), 0) >>> 0;
}
__name(ipv4ToInt, "ipv4ToInt");
function providers(value) {
  if (!value) return [];
  return value.split(",").map((x) => Number.parseInt(x, 10)).filter((n) => Number.isFinite(n));
}
__name(providers, "providers");
function mapAsn(row) {
  const providerAsns = providers(row.provider_asns || null);
  const name = String(row.as_name ?? "");
  return {
    record: {
      id: Number(row.id),
      name,
      countryCode: row.country_code || null,
      countryName: row.country || null,
      asnNumber: Number(row.asn)
    },
    asn: {
      asn: Number(row.asn),
      metadata: {
        handle: name,
        description: name,
        countryCode: row.country_code || "",
        country: row.country || "",
        origin: row.origin || null,
        category: row.category || null,
        registered: row.registered || null,
        lastModified: row.last_modified || null
      },
      stats: {
        ipv4: {
          prefixes: Number(row.prefixes ?? 0),
          prefixesAggregated: Number(row.prefixes_aggregated ?? 0),
          largestPrefix: row.largest_prefix == null ? null : Number(row.largest_prefix),
          totalAddresses: Number(row.total_addresses ?? 0)
        },
        connectivity: { providers: providerAsns.length, providerAsns }
      },
      lastAnnounced: row.last_announced || null
    },
    summary: {
      id: Number(row.id),
      handle: name,
      organization: name,
      countryCode: row.country_code || null,
      country: row.country || null
    }
  };
}
__name(mapAsn, "mapAsn");
async function findByIp(db, ip) {
  const n = ipv4ToInt(ip);
  if (n === null) return null;
  const row = await db.prepare(`${AS_SELECT} WHERE EXISTS (SELECT 1 FROM asn_ipv4 p WHERE p.asn_id = a.id AND ?1 >= p.start_ip AND ?1 <= p.end_ip) ORDER BY s.largest_prefix IS NULL, s.largest_prefix DESC, a.asn ASC LIMIT 1`).bind(n).first();
  return row ? mapAsn(row) : null;
}
__name(findByIp, "findByIp");
async function findCountryByIp(db, ip) {
  const n = ipv4ToInt(ip);
  if (n === null) return null;
  return db.prepare(`SELECT c.country_code AS code, c.country AS name FROM country_ipv4 p JOIN country c ON c.id = p.country_id WHERE ?1 >= p.start_ip AND ?1 <= p.end_ip ORDER BY (p.end_ip - p.start_ip) ASC LIMIT 1`).bind(n).first();
}
__name(findCountryByIp, "findCountryByIp");
async function findByIdent(db, ident) {
  const normalized = ident.trim();
  if (!normalized) return null;
  const num = Number.parseInt(normalized.replace(/^AS/i, ""), 10);
  if (Number.isFinite(num)) {
    const byNum = await db.prepare(`${AS_SELECT} WHERE a.asn = ?1 LIMIT 1`).bind(num).first();
    if (byNum) return mapAsn(byNum);
  }
  const row = await db.prepare(`${AS_SELECT} WHERE LOWER(a."as") LIKE LOWER('%' || ?1 || '%') ORDER BY a.asn ASC LIMIT 1`).bind(normalized).first();
  return row ? mapAsn(row) : null;
}
__name(findByIdent, "findByIdent");
async function logRequest(db, ip, path) {
  if (path === "/api/v1/health" || path === "/api/v1/docs") return;
  try {
    await db.prepare("INSERT INTO api_request (client_ip) VALUES (?1)").bind(ip).run();
  } catch {
  }
}
__name(logRequest, "logRequest");
function normalizeDomain(input) {
  let value = input.trim();
  if (!value) return "";
  try {
    if (value.includes("://")) value = new URL(value).hostname;
  } catch {
  }
  return value.replace(/\.$/, "").toLowerCase();
}
__name(normalizeDomain, "normalizeDomain");
async function doh(name, type) {
  const resp = await fetch(`https://cloudflare-dns.com/dns-query?name=${encodeURIComponent(name)}&type=${type}`, {
    headers: { Accept: "application/dns-json" }
  });
  if (!resp.ok) return [];
  const data = await resp.json();
  const map = { A: 1, NS: 2, CNAME: 5, MX: 15, TXT: 16, AAAA: 28 };
  const wanted = map[type];
  return (data.Answer || []).filter((a) => a.type === wanted).map((a) => a.data.replace(/^"|"$/g, "").replace(/\.$/, ""));
}
__name(doh, "doh");
async function ipInfo(db, ip) {
  const trimmed = ip.trim();
  if (!trimmed) return err("ip is required", 400);
  if (!isIpv4(trimmed) && !(trimmed.includes(":") && /^[0-9a-fA-F:.]+$/.test(trimmed))) return err("invalid ip address", 400);
  if (!isIpv4(trimmed)) return err(`no ASN mapping found for IPv4 address ${trimmed}`, 404);
  const found = await findByIp(db, trimmed);
  if (!found) return err(`no ASN mapping found for IPv4 address ${trimmed}`, 404);
  const geo = await findCountryByIp(db, trimmed);
  return json({ ip: trimmed, asn: found.record.asnNumber, as: found.record.name, country: geo?.name || found.record.countryName || "" });
}
__name(ipInfo, "ipInfo");
async function whois(ipRaw) {
  const ip = ipRaw.trim();
  if (!ip) return err("ip is required", 400);
  const requestURL = `https://rdap.org/ip/${encodeURIComponent(ip)}`;
  const resp = await fetch(requestURL, { headers: { Accept: "application/rdap+json, application/json" } });
  if (resp.status === 404) return err(`no whois/rdap record found for ${ip}`, 404);
  if (resp.status === 400) return err("invalid ip address for rdap", 400);
  if (!resp.ok) return err(`rdap upstream returned HTTP ${resp.status}`, 502);
  const raw = await resp.json();
  const entities = [];
  const walk = /* @__PURE__ */ __name((list) => {
    for (const item of list || []) {
      if (item.handle) entities.push(item.handle);
      if (Array.isArray(item.entities)) walk(item.entities);
    }
  }, "walk");
  walk(raw.entities);
  const remarks = [];
  for (const r of raw.remarks || []) for (const line of r.description || []) if (line.trim()) remarks.push(line.trim());
  let rdapUrl = requestURL;
  for (const link of raw.links || []) {
    if (link.rel?.toLowerCase() === "self" && link.href) {
      rdapUrl = link.href;
      break;
    }
  }
  return json({
    ip,
    handle: raw.handle || "",
    name: raw.name || "",
    type: raw.type || "",
    country: raw.country || "",
    startAddress: raw.startAddress || "",
    endAddress: raw.endAddress || "",
    status: raw.status || [],
    entities,
    events: (raw.events || []).map((e) => ({ action: e.eventAction || "", date: e.eventDate || "" })),
    remarks,
    rdapUrl
  });
}
__name(whois, "whois");
var index_default = {
  async fetch(request, env) {
    const url = new URL(request.url);
    const origin = corsOrigin(env, request.headers.get("Origin") || "");
    if (request.method === "OPTIONS") {
      return new Response(null, {
        status: 204,
        headers: {
          "Access-Control-Allow-Origin": origin,
          "Access-Control-Allow-Methods": "GET,OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type,Authorization",
          "Access-Control-Max-Age": "86400",
          Vary: "Origin"
        }
      });
    }
    if (request.method !== "GET") return withCors(err("method not allowed", 405), origin);
    const path = url.pathname.replace(/\/$/, "") || "/";
    await logRequest(env.DB, clientIp(request), path);
    let res;
    if (path === "/api/v1/health") res = json({ status: "ok", service: "as-ip" });
    else if (path === "/api/v1/docs") {
      res = json({
        service: "as-ip",
        base: "/api/v1",
        endpoints: ["GET /health", "GET /ip/info", "GET /ip/info/:ip", "GET /ip/whois", "GET /ip/whois/:ip", "GET /dns/lookup/*domain", "GET /http/headers", "GET /asn/list", "GET /asn/search/:asn", "GET /asn/to-as/:asn", "GET /as/search/:as", "GET /country/list", "GET /country/search/:country"]
      });
    } else if (path === "/api/v1/http/headers") {
      const omit = /* @__PURE__ */ new Set(["connection", "content-length", "transfer-encoding", "keep-alive", "proxy-connection", "upgrade"]);
      const redact = /* @__PURE__ */ new Set(["cookie", "authorization"]);
      const headers = [];
      for (const [name, value] of request.headers.entries()) {
        const lower = name.toLowerCase();
        if (!omit.has(lower)) headers.push({ name, value: redact.has(lower) ? "[redacted]" : value });
      }
      headers.sort((a, b) => a.name.localeCompare(b.name));
      res = json({ ip: clientIp(request), headers });
    } else if (path === "/api/v1/ip/info") res = await ipInfo(env.DB, clientIp(request));
    else if (path.startsWith("/api/v1/ip/info/")) res = await ipInfo(env.DB, decodeURIComponent(path.slice("/api/v1/ip/info/".length)));
    else if (path === "/api/v1/ip/whois") res = await whois(clientIp(request));
    else if (path.startsWith("/api/v1/ip/whois/")) res = await whois(decodeURIComponent(path.slice("/api/v1/ip/whois/".length)));
    else if (path.startsWith("/api/v1/dns/lookup/")) {
      const domain = normalizeDomain(decodeURIComponent(path.slice("/api/v1/dns/lookup/".length)));
      if (!domain) res = err("domain is required", 400);
      else {
        const [a, aaaa, ns, mx, txt, cname] = await Promise.all([doh(domain, "A"), doh(domain, "AAAA"), doh(domain, "NS"), doh(domain, "MX"), doh(domain, "TXT"), doh(domain, "CNAME")]);
        if (![a, aaaa, ns, mx, txt, cname].some((x) => x.length)) res = err(`no DNS records found for ${domain}`, 404);
        else {
          const addresses = [];
          let asn = 0;
          let asName = "";
          let country = "";
          for (const ip of a) {
            const found = await findByIp(env.DB, ip);
            const geo = await findCountryByIp(env.DB, ip);
            const entry = { ip, asn: found?.record.asnNumber ?? 0, as: found?.record.name ?? "", country: geo?.name || found?.record.countryName || "" };
            addresses.push(entry);
            if (!asn && entry.asn) {
              asn = entry.asn;
              asName = entry.as;
              country = entry.country;
            }
          }
          res = json({
            domain,
            a,
            aaaa,
            ns,
            mx: mx.map((value) => {
              const [pref, host] = value.split(/\s+/, 2);
              return { host: host || value, pref: Number.parseInt(pref, 10) || 0 };
            }),
            txt,
            cname: cname[0] || "",
            asn,
            as: asName,
            country,
            addresses
          });
        }
      }
    } else if (path === "/api/v1/asn/list") {
      const { results } = await env.DB.prepare('SELECT asn AS number, "as" AS name FROM asn ORDER BY asn').all();
      res = json((results || []).map((r) => ({ asn: r.number, as: r.name })));
    } else if (path.startsWith("/api/v1/asn/search/")) {
      const found = await findByIdent(env.DB, decodeURIComponent(path.slice("/api/v1/asn/search/".length)));
      res = found ? json(found.asn) : err("ASN not found", 404);
    } else if (path.startsWith("/api/v1/asn/to-as/")) {
      const found = await findByIdent(env.DB, decodeURIComponent(path.slice("/api/v1/asn/to-as/".length)));
      res = found ? json(found.summary) : err("ASN not found", 404);
    } else if (path.startsWith("/api/v1/as/search/")) {
      const found = await findByIdent(env.DB, decodeURIComponent(path.slice("/api/v1/as/search/".length)));
      res = found ? json(found.asn) : err("AS not found", 404);
    } else if (path === "/api/v1/country/list") {
      const { results } = await env.DB.prepare("SELECT country_code AS code, country AS name FROM country ORDER BY country_code").all();
      res = json((results || []).map((r) => ({ code: r.code, name: r.name })));
    } else if (path.startsWith("/api/v1/country/search/")) {
      const code = decodeURIComponent(path.slice("/api/v1/country/search/".length)).trim();
      if (!/^[a-zA-Z]{2}$/.test(code)) res = err("country must be an ISO 3166-1 alpha-2 code", 400);
      else {
        const { results } = await env.DB.prepare(`${AS_SELECT} WHERE LOWER(c.country_code) = LOWER(?1) ORDER BY a.asn ASC`).bind(code.toUpperCase()).all();
        res = json((results || []).map((row) => mapAsn(row).asn));
      }
    } else res = err("not found", 404);
    return withCors(res, origin);
  }
};
export {
  index_default as default
};
//# sourceMappingURL=index.js.map
