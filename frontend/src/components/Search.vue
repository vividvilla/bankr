<template>
	<div id="search" class="search">
		<div class="wrapper">
			<!--Search header-->
			<div class="header">
				<!--Logo-->
				<h1 class="logo">bankr</h1>

				<!--Search input-->
				<div class="search-box">
					<span class="icon icon-search"></span>
					<input class="search-input"
						placeholder="Find any bank name, address, IFSC etc."
						v-model="searchTerm" @keyup.enter="doSearch()"/>
					<progress-bar></progress-bar>
				</div>
			</div>

			<!--Search results-->
			<div class="search-results" v-if="results">
				<!--Query time-->
				<div class="query-time" v-if="results && results.results.length > 0">
					Took {{ results.took }} to search
				</div>

				<!--Search result items-->
				<transition-group name="list" tag="div" appear>
					<div class="search-item" v-for="result in results.results" :key="result.id">
						<!--Top row-->
						<div class="search-item-header">
							<h2 class="name">{{ result.fields.name }}</h2>
							<span class="ifsc">{{ result.fields.IFSC }}</span>
							<span>&nbsp;</span>
						</div>

						<!--Second row-->
						<div class="search-item-body">
							<span class="branch">{{ result.fields.branch }}</span>
							<div class="info">
								<div class="micr" v-if="result.fields.MICR">
									<span class="icon icon-qrcode"></span>
									<span>{{ result.fields.MICR }}</span>
								</div>
								<div class="phone-number">
									<span class="icon icon-phone"></span>
									<span>{{ result.fields.contact.split(".")[0] }}</span>
								</div>
							</div>
						</div>

						<!--Last row-->
						<div class="search-item-footer">
							<div class="address">
								<span class="icon icon-location"></span><span>{{ result.fields.address }}</span>
							</div>
						</div>
					</div>
				</transition-group>
			</div>
		</div>
	</div>
</template>

<script>
	import Vue from "vue"
	import axios from "axios"

	import ProgressBar from "./ProgressBar"

	export default Vue.extend({
		components: {
			"progress-bar": ProgressBar
		},
		data () {
			return {
				searchTerm: "",
				results: null
			}
		},
		mounted () {
			if (this.$route && this.$route.query && this.$route.query.q) {
				this.searchTerm = this.$route.query.q
				this.$nextTick(() => this.doSearch())
			}
		},
		created () {
		},
		methods: {
			startProgress () {
				this.$events.$emit("startProgress", true)
			},
			successProgress () {
				this.$events.$emit("finishProgress", {success: true})
			},
			errorProgress () {
				this.$events.$emit("finishProgress", {success: true})
			},
			doSearch () {
				// Set current term as query param
				this.$router.push({ path: "/", query: { q: this.searchTerm } })

				// start progress bar
				this.startProgress()

				// prepare the params
				let params = {
					q: this.searchTerm
				}

				// API call
				axios.get("http://127.0.0.1:3000/api/search", { params })
					.then((response) => {
						this.results = response.data
						this.successProgress()
					})
					.catch((error) => {
						console.log(error)
						this.errorProgress()
					})
			}
		}
	})
</script>
