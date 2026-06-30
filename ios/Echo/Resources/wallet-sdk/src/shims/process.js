// Minimal Node `process` shim with a real `env`, replacing jspm's polyfill
// whose default export is undefined under esbuild (process$1.env -> throw).
const proc = {
  env: {},
  browser: true,
  version: '',
  versions: {},
  platform: 'browser',
  argv: [],
  nextTick: (cb, ...args) => setTimeout(() => cb(...args), 0),
  cwd: () => '/',
  on: () => {},
  once: () => {},
  off: () => {},
  emit: () => {},
};
export default proc;
export const env = proc.env;
export const browser = true;
export const nextTick = proc.nextTick;
