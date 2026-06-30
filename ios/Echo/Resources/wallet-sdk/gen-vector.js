// Cross-validation vector generator.
//
// Runs the BUILT bundle (echo-wallet.bundle.js) inside a bare vm sandbox that
// approximates JavaScriptCore (no Node globals: no require/process/Buffer), to
// (a) prove the bundle is self-contained, and (b) emit a deterministic vector
// the Go test re-derives/verifies. Uses a fixed BIP-39 mnemonic so no runtime
// randomness is needed (importMnemonic + sign are deterministic).
//
// NOTE: bare JavaScriptCore also lacks crypto.getRandomValues; key *generation*
// in the app will require Swift to inject it. This vector path deliberately
// avoids generation to isolate the deterministic crypto.

const fs = require('fs');
const path = require('path');
const vm = require('vm');

const bundle = fs.readFileSync(path.join(__dirname, 'echo-wallet.bundle.js'), 'utf8');

// Minimal globals a real JSC context has (plus what we inject from Swift).
// Globals a real JSC context must be given by Swift (M1 injects the same set).
const sandbox = {
  console,
  TextEncoder,
  TextDecoder,
  setTimeout,
  clearTimeout,
  process: { env: {}, browser: true, version: '', versions: {}, nextTick: (cb) => setTimeout(cb, 0) },
};
sandbox.globalThis = sandbox;
vm.createContext(sandbox);
vm.runInContext(bundle, sandbox, { filename: 'echo-wallet.bundle.js' });

const EchoWallet = sandbox.EchoWallet;
if (!EchoWallet) {
  console.error('FAIL: EchoWallet global not defined by bundle');
  process.exit(1);
}

const MNEMONIC =
  'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about';
const MESSAGE = 'echo-wallet-ownership:did:key:zVECTOR:nonce-0123456789';

(async () => {
  const acct = EchoWallet.importMnemonic(MNEMONIC);
  const signature = await EchoWallet.signMessage(acct.privateKey, MESSAGE);
  const vector = {
    mnemonic: MNEMONIC,
    privateKey: acct.privateKey,
    publicKey: acct.publicKey,
    address: acct.address,
    message: MESSAGE,
    signature,
  };
  const out = path.join(__dirname, 'testvector.json');
  fs.writeFileSync(out, JSON.stringify(vector, null, 2) + '\n');
  console.log(JSON.stringify(vector, null, 2));
  console.error('wrote ' + out);
})().catch((e) => {
  console.error('FAIL:', e && e.stack ? e.stack : e);
  process.exit(1);
});
