/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // The portal is a thin client. All compliance data is fetched server-side
  // from the Go Comply API; never proxy raw PII to the browser.
  env: {
    NEXT_PUBLIC_APP_NAME: "Echo Comply",
  },
};

export default nextConfig;
