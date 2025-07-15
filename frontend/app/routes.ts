import { type RouteConfig, index } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  { path: "/screener", file: "routes/screener.tsx" },
] satisfies RouteConfig;
