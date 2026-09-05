import fs from 'fs'

const scriptPath = 'C:/Users/armin/GitHub/asip/asip-api/cloudflare/.cf-bundle/index.js'
const outPath = 'C:/Users/armin/GitHub/asip/asip-api/cloudflare/.cf-bundle/mcp-deploy-code.js'
const b64 = fs.readFileSync(scriptPath).toString('base64')

const code = `async () => {
  const b64 = ${JSON.stringify(b64)};
  const bin = Uint8Array.from(atob(b64), c => c.charCodeAt(0));
  const script = new TextDecoder().decode(bin);
  const metadata = {
    main_module: 'index.js',
    compatibility_date: '2026-08-31',
    compatibility_flags: ['nodejs_compat'],
    bindings: [
      { type: 'd1', name: 'DB', id: '0bd20123-61ba-494e-ada1-27eb4fdb971a' },
      { type: 'plain_text', name: 'CORS_ORIGINS', text: 'http://localhost:5173,http://127.0.0.1:5173,https://asip.armindashti.workers.dev' }
    ]
  };
  const b = '----FormBoundary' + Date.now();
  const body = [
    '--' + b,
    'Content-Disposition: form-data; name="metadata"',
    'Content-Type: application/json',
    '',
    JSON.stringify(metadata),
    '--' + b,
    'Content-Disposition: form-data; name="index.js"; filename="index.js"',
    'Content-Type: application/javascript+module',
    '',
    script,
    '--' + b + '--',
    ''
  ].join('\\r\\n');
  const put = await cloudflare.request({
    method: 'PUT',
    path: '/accounts/' + accountId + '/workers/scripts/asip-api',
    body,
    contentType: 'multipart/form-data; boundary=' + b,
    rawBody: true
  });
  const sub = await cloudflare.request({
    method: 'POST',
    path: '/accounts/' + accountId + '/workers/scripts/asip-api/subdomain',
    body: { enabled: true }
  });
  return {
    putSuccess: put.success,
    putErrors: put.errors,
    putStatus: put.status,
    putMessages: put.messages,
    subSuccess: sub.success,
    subErrors: sub.errors
  };
}`

fs.writeFileSync(outPath, code)
console.log('wrote', outPath, 'chars', code.length)
