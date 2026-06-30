// Node `crypto` shim: only the surface dag4-keystore touches (randomBytes,
// createHash, createHmac, pbkdf2). Backed by @noble; no streams.
import createHash from './create-hash.js';
import createHmac from './create-hmac.js';
import randomBytes from './randombytes.js';
import { pbkdf2, pbkdf2Sync } from './pbkdf2.js';

export { createHash, createHmac, randomBytes, pbkdf2, pbkdf2Sync };
export default { createHash, createHmac, randomBytes, pbkdf2, pbkdf2Sync };
