// Pure-JS `create-hmac` over @noble/hashes (BIP32 HMAC-SHA512, etc.).
// CJS: consumed via __toCommonJS(...) and called as a function (see create-hash).
const { hmac } = require('@noble/hashes/hmac');
const { sha256 } = require('@noble/hashes/sha256');
const { sha512 } = require('@noble/hashes/sha512');
const { Buffer } = require('buffer');

const ALGOS = { sha256, sha512 };

function toBytes(data, enc) {
  if (data == null) return new Uint8Array(0);
  if (typeof data === 'string') return Uint8Array.from(Buffer.from(data, enc || 'utf8'));
  return data instanceof Uint8Array ? data : Uint8Array.from(data);
}

module.exports = function createHmac(algorithm, key) {
  const algo = ALGOS[String(algorithm).toLowerCase()] || sha512;
  const h = hmac.create(algo, toBytes(key));
  return {
    update(data, enc) {
      h.update(toBytes(data, enc));
      return this;
    },
    digest(enc) {
      const out = Buffer.from(h.digest());
      return enc ? out.toString(enc) : out;
    },
  };
};
