// Compiles contracts/EchoAnchor.sol to build/EchoAnchor.{abi,bin} with the JavaScript build of solc.
// (The ethereum/solc Docker image is amd64 only.) Target EVM "paris" so the bytecode runs on any
// fork the private network is configured with.
const fs = require('fs'), path = require('path'), solc = require('solc');
const src = fs.readFileSync(path.join(__dirname, 'contracts/EchoAnchor.sol'), 'utf8');
const out = JSON.parse(solc.compile(JSON.stringify({
  language: 'Solidity',
  sources: { 'EchoAnchor.sol': { content: src } },
  settings: { evmVersion: 'paris', optimizer: { enabled: true, runs: 200 },
              outputSelection: { '*': { '*': ['abi', 'evm.bytecode.object'] } } },
})));
const errors = (out.errors || []).filter(e => e.severity === 'error');
if (errors.length) { console.error(errors.map(e => e.formattedMessage).join('\n')); process.exit(1); }
const c = out.contracts['EchoAnchor.sol'].EchoAnchor;
fs.mkdirSync(path.join(__dirname, 'build'), { recursive: true });
fs.writeFileSync(path.join(__dirname, 'build/EchoAnchor.abi'), JSON.stringify(c.abi));
fs.writeFileSync(path.join(__dirname, 'build/EchoAnchor.bin'), c.evm.bytecode.object);
console.log('compiled EchoAnchor with solc', solc.version(), '-', c.evm.bytecode.object.length / 2, 'bytes');
