// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from "vue"
import ga from "vue-ga"
import router from "./router"
import VueRouter from "vue-router"
import VueClipboards from "./directives/vue-clipboards"

// Disable production tip
Vue.config.productionTip = false

// Event bus to manage cross component communications
let eventBus = new Vue()
Vue.prototype.$events = eventBus

Vue.use(VueRouter)
Vue.use(VueClipboards)

// Google analytics
if (process.env.GOOGLE_ANALYTICS_ID) {
	ga(router, process.env.GOOGLE_ANALYTICS_ID)
}

/* eslint-disable no-new */
new Vue({
	el: "#app",
	template: "<router-view></router-view>",
	router
})
