// Compiles every contract in chain/contracts to internal/besu/contract/<Name>.{abi,bin}.
// Those artefacts are committed and embedded in the gateway, so building Echo needs no Solidity
// toolchain. `make contracts-check` fails if they do not match the sources.
//   node compile.js          write artefacts
//   node compile.js --check  compare only; exit 1 on any difference
const fs = require('fs'), path = require('path'), solc = require('solc');
const srcDir = path.join(__dirname, 'contracts');
const outDir = path.join(__dirname, '..', 'internal', 'besu', 'contract');
const check = process.argv.includes('--check');
const sources = {};
for (const f of fs.readdirSync(srcDir).filter(f => f.endsWith('.sol')))
  sources[f] = { content: fs.readFileSync(path.join(srcDir, f), 'utf8') };
const out = JSON.parse(solc.compile(JSON.stringify({
  language: 'Solidity', sources,
  // "paris": no opcode newer than the forks the private network enables.
  // No metadata hash, so the bytecode depends on the source and compiler version alone.
  settings: { evmVersion: 'paris', optimizer: { enabled: true, runs: 200 },
              metadata: { bytecodeHash: 'none', appendCBOR: false },
              outputSelection: { '*': { '*': ['abi', 'evm.bytecode.object'] } } },
})));
const errors = (out.errors || []).filter(e => e.severity === 'error');
if (errors.length) { console.error(errors.map(e => e.formattedMessage).join('\n')); process.exit(1); }
let stale = 0;
for (const file of Object.keys(out.contracts)) for (const [name, c] of Object.entries(out.contracts[file])) {
  for (const [ext, body] of [['abi', JSON.stringify(c.abi)], ['bin', c.evm.bytecode.object]]) {
    const p = path.join(outDir, `${name}.${ext}`);
    if (check) {
      if (!fs.existsSync(p) || fs.readFileSync(p, 'utf8') !== body) { console.error(`stale: ${p}`); stale++; }
    } else fs.writeFileSync(p, body);
  }
  console.log(`${check ? 'checked' : 'compiled'} ${name} (solc ${solc.version()}, ${c.evm.bytecode.object.length / 2} bytes)`);
}
process.exit(stale ? 1 : 0);
