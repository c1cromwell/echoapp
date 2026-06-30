// Bundles src/index.js into a single self-contained IIFE for JavaScriptCore.
// No runtime require(); Node builtins (crypto/stream/events/buffer) are
// polyfilled in-bundle. assert/util use minimal local shims (jspm's polyfills
// throw at init in a bare context). brotli-wasm is shimmed for M0 (the tx
// serialization that needs it is wired in M3). `process` is a host global that
// Swift injects into the JSContext (gen-vector.js mirrors it).
const esbuild = require('esbuild');
const { polyfillNode } = require('esbuild-plugin-polyfill-node');
const path = require('path');

const shim = (name) => path.join(__dirname, 'src', 'shims', name);

// Runs BEFORE polyfillNode so our shims win. Intercepts the buggy jspm
// assert/util/process polyfills both as bare specifiers AND via jspm's internal
// relative imports (e.g. ./process.js from inside @jspm/core/nodelibs/browser).
const shimResolver = {
  name: 'echo-shim-resolver',
  setup(build) {
    const map = {
      assert: shim('assert.js'),
      util: shim('util.js'),
      process: shim('process.js'),
      crypto: shim('crypto.js'),
      stream: shim('stream.js'),
      'create-hash': shim('create-hash.js'),
      'create-hmac': shim('create-hmac.js'),
      pbkdf2: shim('pbkdf2.js'),
      randombytes: shim('randombytes.js'),
      'brotli-wasm': shim('brotli-wasm.js'),
    };
    const bare = new RegExp('^(node:)?(' + Object.keys(map).join('|') + ')$');
    build.onResolve({ filter: bare }, (args) => ({
      path: map[args.path.replace(/^node:/, '')],
    }));
    // uuid (bare and /v4 subpath) -> host-RNG shim (drops the uuid dep).
    build.onResolve({ filter: /^uuid(\/v4)?$/ }, () => ({ path: shim('uuid-v4.js') }));
    // jspm nodelibs import each other relatively; redirect those to our shims.
    build.onResolve({ filter: /^\.\.?\// }, (args) => {
      if (!args.importer.includes('@jspm/core/nodelibs/browser')) return undefined;
      const base = path.basename(args.path).replace(/\.js$/, '');
      return map[base] ? { path: map[base] } : undefined;
    });
  },
};

esbuild
  .build({
    entryPoints: [path.join(__dirname, 'src', 'index.js')],
    bundle: true,
    format: 'iife',
    globalName: 'EchoWalletBundle',
    platform: 'browser',
    target: ['es2020'],
    outfile: path.join(__dirname, 'echo-wallet.bundle.js'),
    define: { 'process.env.NODE_ENV': '"production"', global: 'globalThis' },
    plugins: [
      shimResolver,
      // crypto/stream are handled by our shims above; let polyfillNode cover any
      // remaining leftovers (events, etc.) and the Buffer global.
      polyfillNode({ globals: { process: false, buffer: true }, polyfills: { crypto: false } }),
    ],
    logLevel: 'info',
  })
  .then(() => console.log('built echo-wallet.bundle.js'))
  .catch((e) => {
    console.error(e);
    process.exit(1);
  });
