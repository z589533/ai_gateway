const { VITE_HIDE_HOME } = import.meta.env;
const Layout = () => import("@/layout/index.vue");

export default {
  path: "/",
  name: "GatewayRoot",
  component: Layout,
  redirect: "/gateway/index",
  meta: {
    icon: "ep/monitor",
    title: "AI Gateway",
    rank: 0
  },
  children: [
    {
      path: "/gateway/index",
      name: "GatewayAdmin",
      component: () => import("@/views/gateway/index.vue"),
      meta: {
        title: "管理台",
        showLink: VITE_HIDE_HOME === "true" ? false : true
      }
    }
  ]
} satisfies RouteConfigsTable;
