// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from "vue"
import App from "./App"

// Disable production tip
Vue.config.productionTip = false

// Event bus to manage cross component communications
let eventBus = new Vue()
Vue.prototype.$events = eventBus

/* eslint-disable no-new */
new Vue({
	el: "#app",
	template: "<App/>",
	components: { App }
})
