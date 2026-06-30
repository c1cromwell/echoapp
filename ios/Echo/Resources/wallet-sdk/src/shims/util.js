// Minimal Node `util` shim (replaces jspm's buggy polyfill). Only the members
// ethereumjs-util / readable-stream actually reference.
export function inherits(ctor, superCtor) {
  ctor.super_ = superCtor;
  Object.setPrototypeOf(ctor.prototype, superCtor.prototype);
}
export function inspect(x) {
  try {
    return typeof x === 'string' ? x : JSON.stringify(x);
  } catch (_) {
    return String(x);
  }
}
export function debuglog() {
  return function () {};
}
export function deprecate(fn) {
  return fn;
}
export function promisify(fn) {
  return (...args) =>
    new Promise((resolve, reject) => {
      fn(...args, (err, val) => (err ? reject(err) : resolve(val)));
    });
}
export function format(fmt, ...args) {
  let i = 0;
  return String(fmt).replace(/%[sdj%]/g, (m) => (m === '%%' ? '%' : String(args[i++])));
}
export function isBuffer(v) {
  return v != null && v._isBuffer === true;
}
export const types = {
  isAnyArrayBuffer: (v) => v instanceof ArrayBuffer,
  isUint8Array: (v) => v instanceof Uint8Array,
};
export default { inherits, inspect, debuglog, deprecate, promisify, format, isBuffer, types };
