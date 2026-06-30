// M0 shim: brotli-wasm is only needed for transaction-v2 BROTLI serialization
// (used by token-lock / delegated-stake signing in M3). M0 only exercises
// keygen / address / message-signing, none of which compress. We stub it so the
// import graph resolves without a .wasm loader. M3 replaces this with the real
// brotli-wasm web build wired for JavaScriptCore.
function notWired() {
  throw new Error('brotli-wasm not wired in M0 bundle (tx serialization is M3)');
}
const stub = { compress: notWired, decompress: notWired };
export default Promise.resolve(stub);
export const compress = notWired;
export const decompress = notWired;
