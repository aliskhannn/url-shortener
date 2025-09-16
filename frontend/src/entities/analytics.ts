export interface Analytics {
  id: string           // uuid
  alias: string        // short alias
  userAgent: string    // raw user agent
  device: string       // device type
  os: string           // operating system
  browser: string      // browser name
  ip: string           // client ip
  createdAt: string    // timestamp в ISO формате
}