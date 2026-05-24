import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'https://audit-bot-bok1.onrender.com/api/:path*',
      },
    ];
  },
};
export default nextConfig;