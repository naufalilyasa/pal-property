import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "res.cloudinary.com",
        pathname: "/**",
      },
      {
        protocol: "https",
        hostname: "images.example",
        pathname: "/**",
      },
    ],
    qualities: [60, 75, 90],
  },
};

export default nextConfig;
