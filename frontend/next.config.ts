import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  reactCompiler: true,
  cacheComponents: true,
  partialPrefetching: true,
  experimental: {
    turbopackGc: true,
    turbopackRustReactCompiler: true,
    turbopackLazyDynamicImports: true,
    turbopackPluginRuntimeStrategy: 'workerThreads',
  },
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
