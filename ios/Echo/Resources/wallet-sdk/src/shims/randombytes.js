// Secure RNG shim. In JavaScriptCore there is no global crypto.getRandomValues
// by default — Swift injects one (SecRandomCopyBytes) before loading the bundle.
// CJS: consumed via __toCommonJS(...) and called as a function.
const { Buffer } = require('buffer');

module.exports = function randomBytes(n, cb) {
  const out = new Uint8Array(n);
  if (globalThis.crypto && typeof globalThis.crypto.getRandomValues === 'function') {
    globalThis.crypto.getRandomValues(out);
  } else {
    throw new Error('no secure RNG: host must provide crypto.getRandomValues');
  }
  const buf = Buffer.from(out);
  if (cb) {
    cb(null, buf);
    return;
  }
  return buf;
};
