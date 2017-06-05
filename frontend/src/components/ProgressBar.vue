<template>
	<div id="prog-bar" class="progress">
		<div class="progress-bar"></div>
	</div>
</template>

<script>
	import Vue from "vue"

	export default Vue.extend({
		props: {
		},
		data () {
			return {
				width: 0,
				barColor: "#0070e0",
				barSccessColor: "#0070e0",
				barErrorColor: "#fc677e",
				timer: null,
				barElement: null
			}
		},
		created () {
			this.$events.$on("startProgress", this.startProgress)
			this.$events.$on("finishProgress", this.finishProgress)

			// Reset on start
			this.$nextTick(() => {
				this.resetBarElement()
			})
		},
		methods: {
			resetBarElement () {
				this.barElement = this.$el.querySelector(".progress-bar")
				this.barElement.style.background = "#64cbfc"
				this.barElement.style.width = "0"
				this.barElement.hidden = true
			},
			setWidth (width) {
				if (width < 100) {
					this.width = width
				} else {
					this.width = 100
				}
			},
			updateProgress () {
				this.barElement.style.width = this.width + "%"
			},
			startProgress () {
				this.barElement.hidden = false

				// Initialize with 30% width
				this.width = 30
				this.updateProgress()

				// Set timer
				this.timer = setInterval(() => {
					this.setWidth(this.width + 10)
					this.updateProgress()
				}, 500)
			},
			finishProgress (options) {
				if (options.error) this.barElement.style.background = this.barErrorColor
				if (options.success) this.barElement.style.background = this.barSccessColor

				this.clearTimer()

				// set width to 100%
				this.setWidth(100)
				this.updateProgress()

				// After a second make reset the element
				setTimeout(() => {
					// this.barElement.hidden = true
					this.resetProgress()
				}, 500)
			},
			resetProgress () {
				this.clearTimer()
				this.resetBarElement()
			},
			clearTimer () {
				if (this.timer) {
					clearInterval(this.timer)
					this.timer = null
				}
			}
		}
	})
</script>
