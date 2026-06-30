// Pure-JS `pbkdf2` package shim over @noble/hashes (BIP-39 seed derivation).
import { pbkdf2 as noblePbkdf2 } from '@noble/hashes/pbkdf2';
import { sha256 } from '@noble/hashes/sha256';
import { sha512 } from '@noble/hashes/sha512';
import { sha1 } from '@noble/hashes/sha1';
import { Buffer } from 'buffer';

const ALGOS = { sha256, sha512, sha1 };

function run(pass, salt, iters, keylen, digest) {
  const algo = ALGOS[String(digest || 'sha1').toLowerCase()] || sha256;
  const out = noblePbkdf2(algo, toBytes(pass), toBytes(salt), { c: iters, dkLen: keylen });
  return Buffer.from(out);
}
function toBytes(d) {
  if (d == null) return new Uint8Array(0);
  if (typeof d === 'string') return Uint8Array.from(Buffer.from(d, 'utf8'));
  return d instanceof Uint8Array ? d : Uint8Array.from(d);
}

export function pbkdf2Sync(pass, salt, iters, keylen, digest) {
  return run(pass, salt, iters, keylen, digest);
}
export function pbkdf2(pass, salt, iters, keylen, digest, cb) {
  try {
    cb(null, run(pass, salt, iters, keylen, digest));
  } catch (e) {
    cb(e);
  }
}
export default { pbkdf2, pbkdf2Sync };
