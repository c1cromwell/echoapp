// Minimal `stream` shim. After create-hash/create-hmac are shimmed, nothing in
// the keygen/address/sign path actually instantiates a stream; these classes
// exist only so leftover `import { Transform } from 'stream'` resolves.
class Stream {}
class Transform {
  constructor() {}
}
class Readable {}
class Writable {}
class Duplex {}
class PassThrough {}
export { Stream, Transform, Readable, Writable, Duplex, PassThrough };
export default Stream;
