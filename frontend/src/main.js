// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from "vue"
import VueRouter from "vue-router"

import router from "./router"

// Disable production tip
Vue.config.productionTip = false

// Event bus to manage cross component communications
let eventBus = new Vue()
Vue.prototype.$events = eventBus

Vue.use(VueRouter)

/* eslint-disable no-new */
new Vue({
	el: "#app",
	template: "<router-view></router-view>",
	router
})
