// Echo wallet SDK — the only JS that runs inside the iOS JavaScriptCore context.
//
// It exposes a tiny, synchronous-friendly surface over dag4's keyStore so Swift
// can generate/import keys, derive the DAG address, and sign. Key material is
// passed in per call by Swift (held in the Keychain) and never persisted here.
//
// Built into a single self-contained file by build.js (esbuild, IIFE) so it has
// zero require()/Node deps at runtime in JSC.

import { Buffer } from 'buffer';
// JavaScriptCore lacks Node globals dag4 transitively expects.
globalThis.Buffer = globalThis.Buffer || Buffer;

import { dag4 } from '@stardust-collective/dag4';

const keyStore = dag4.keyStore;

// Promisify: keyStore.sign is async; expose a callback the Swift bridge awaits.
function toHex(x) {
  return typeof x === 'string' ? x : Buffer.from(x).toString('hex');
}

const EchoWallet = {
  // Returns a fresh BIP-39 mnemonic (DAG wallet seed).
  generateMnemonic() {
    return keyStore.generateSeedPhrase();
  },

  // Derives the canonical DAG account from a mnemonic.
  // -> { privateKey, publicKey, address }
  importMnemonic(mnemonic) {
    const privateKey = keyStore.getPrivateKeyFromMnemonic(mnemonic);
    const publicKey = keyStore.getPublicKeyFromPrivate(privateKey);
    const address = keyStore.getDagAddressFromPublicKey(publicKey);
    return { privateKey, publicKey, address };
  },

  // -> { publicKey, address } for a raw private key.
  accountFromPrivateKey(privateKey) {
    const publicKey = keyStore.getPublicKeyFromPrivate(privateKey);
    const address = keyStore.getDagAddressFromPublicKey(publicKey);
    return { publicKey, address };
  },

  deriveAddress(publicKeyHex) {
    return keyStore.getDagAddressFromPublicKey(publicKeyHex);
  },

  // Proof-of-ownership / generic message signing: secp256k1 over SHA512(msg),
  // DER hex. Verified server-side by the Go secp256k1 stack.
  // Returns a Promise<string> (DER hex signature).
  signMessage(privateKey, message) {
    return keyStore.sign(privateKey, message).then(toHex);
  },
};

globalThis.EchoWallet = EchoWallet;
export default EchoWallet;
