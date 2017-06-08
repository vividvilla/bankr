import VueRouter from "vue-router"

import App from "./App"

const routes = [
  { path: "/", component: App }
]

const router = new VueRouter({ mode: "history", routes })

export default router
