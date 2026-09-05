import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  transpilePackages: ["@hola/api-client"],
  agentRules: false,
};

export default nextConfig;
