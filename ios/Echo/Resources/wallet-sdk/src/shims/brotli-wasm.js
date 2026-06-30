// brotli-wasm for bare JavaScriptCore.
//
// The published web build fetches the .wasm by URL (import.meta.url / fetch),
// neither of which exists in JSC. wasm-bindgen's init() also accepts the wasm
// BYTES directly, so we embed the .wasm (base64 via esbuild's loader) and pass
// it in — no fetch, no import.meta.
//
// dag4's serializeBrotli (with `self` undefined, as in JSC) takes the Node
// branch: require('brotli-wasm') then module.compress(bytes, { quality: 2 }).
// We expose `compress`/`decompress` directly and make them async so wasm init
// can complete first (dag4 awaits the result). We deliberately do NOT define a
// global `self`, which would route @noble through an unavailable crypto.subtle.
import init, * as brotliWasm from '../../node_modules/brotli-wasm/pkg.web/brotli_wasm.js';
import wasmBase64 from '../../node_modules/brotli-wasm/pkg.web/brotli_wasm_bg.wasm';
import { Buffer } from 'buffer';

const wasmBytes = Uint8Array.from(Buffer.from(wasmBase64, 'base64'));

let readyPromise = null;
function ready() {
  if (!readyPromise) readyPromise = init(wasmBytes).then(() => brotliWasm);
  return readyPromise;
}

export async function compress(bytes, options) {
  const mod = await ready();
  return mod.compress(bytes, options);
}
export async function decompress(bytes) {
  const mod = await ready();
  return mod.decompress(bytes);
}

export default { compress, decompress };
