// uuid v4 over the host RNG (crypto.getRandomValues, injected by Swift),
// replacing the `uuid` package — its patched majors (>=11.1.1) drop the
// `uuid/v4` subpath dag4-keystore imports, so we shim it instead. Also removes
// uuid from the shipped bundle. CJS: dag4 does `import v4 from 'uuid/v4'`.
const HEX = [];
for (let i = 0; i < 256; i++) HEX.push((i + 0x100).toString(16).slice(1));

function v4() {
  const b = new Uint8Array(16);
  if (globalThis.crypto && typeof globalThis.crypto.getRandomValues === 'function') {
    globalThis.crypto.getRandomValues(b);
  } else {
    throw new Error('no secure RNG for uuid: host must provide crypto.getRandomValues');
  }
  b[6] = (b[6] & 0x0f) | 0x40; // version 4
  b[8] = (b[8] & 0x3f) | 0x80; // variant
  return (
    HEX[b[0]] + HEX[b[1]] + HEX[b[2]] + HEX[b[3]] + '-' +
    HEX[b[4]] + HEX[b[5]] + '-' +
    HEX[b[6]] + HEX[b[7]] + '-' +
    HEX[b[8]] + HEX[b[9]] + '-' +
    HEX[b[10]] + HEX[b[11]] + HEX[b[12]] + HEX[b[13]] + HEX[b[14]] + HEX[b[15]]
  );
}

module.exports = v4;
module.exports.default = v4;
module.exports.v4 = v4;
