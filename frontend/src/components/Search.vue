<template>
	<div id="search" class="search">
		<div class="wrapper">
			<div class="header">
				<h1 class="logo">bankr</h1>
				<div class="search-box">
					<span class="icon icon-search"></span>
					<input class="search-input"
						placeholder="Find any bank name, address, IFSC etc."
						v-model="searchTerm" @keyup.enter="doSearch()"/>
					<progress-bar></progress-bar>
				</div>
			</div>
			<div class="search-results" v-if="this.results">
				<div class="search-item" v-for="result in results.results">
					<div class="search-item-header">
						<h2 class="name">{{ result.fields.name }}</h2>
						<span class="ifsc">{{ result.fields.IFSC }}</span>
					</div>
					<div class="search-item-body">
						<div class="branch">
							<span></span> {{ result.fields.branch }}
						</div>
						<div class="address">
							<span></span> {{ result.fields.address }}
						</div>
					</div>
				</div>
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
				this.startProgress()

				let params = {
					q: this.searchTerm
				}

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
