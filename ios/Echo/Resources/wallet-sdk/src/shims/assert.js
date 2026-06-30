// Minimal Node `assert` shim. ethereumjs-util uses assert() for sanity checks;
// the jspm polyfill drags in a buggy util/process chain, so we replace it.
function assert(value, message) {
  if (!value) throw new Error(message || 'assertion failed');
}
assert.ok = assert;
assert.equal = (a, b, m) => {
  if (a != b) throw new Error(m || 'assert.equal failed');
};
assert.strictEqual = (a, b, m) => {
  if (a !== b) throw new Error(m || 'assert.strictEqual failed');
};
export default assert;
export { assert };
