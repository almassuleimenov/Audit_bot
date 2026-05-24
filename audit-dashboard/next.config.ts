import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'https://audit-bot-1-j1xj.onrender.com/api/:path*',
      },
    ];
  },
};
export default nextConfig;