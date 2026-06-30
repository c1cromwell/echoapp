// Pure-JS `create-hash` over @noble/hashes — replaces the create-hash →
// cipher-base → stream (jspm) chain that breaks in a bare JS context.
// CJS on purpose: ethereum-cryptography's hdkey-crypto shim does
// __toCommonJS(require('create-hash')) and calls the result as a function, so
// module.exports must BE the factory function (not an ESM namespace).
const { sha256 } = require('@noble/hashes/sha256');
const { sha512 } = require('@noble/hashes/sha512');
const { ripemd160 } = require('@noble/hashes/ripemd160');
const { sha1 } = require('@noble/hashes/sha1');
const { Buffer } = require('buffer');

const ALGOS = { sha256, sha512, ripemd160, rmd160: ripemd160, sha1 };

function toBytes(data, enc) {
  if (data == null) return new Uint8Array(0);
  if (typeof data === 'string') return Uint8Array.from(Buffer.from(data, enc || 'utf8'));
  return data instanceof Uint8Array ? data : Uint8Array.from(data);
}

module.exports = function createHash(algorithm) {
  const algo = ALGOS[String(algorithm).toLowerCase()] || sha256;
  const h = algo.create();
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
