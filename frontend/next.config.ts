import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  reactCompiler: true,
  experimental: {
    turbopackGc: true,
    turbopackRustReactCompiler: true,
    turbopackLazyDynamicImports: true,
  },
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
