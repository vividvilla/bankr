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
						placeholder="Find any bank by name, address, IFSC etc."
						v-model="searchTerm" @keyup.enter="doSearch(searchTerm)"
						:class="{ error: inputError }"/>
					<progress-bar></progress-bar>
				</div>

				<div class="user-location">
					<transition name="fade">
						<div class="checking-location" v-if="isCheckingNavigation">
							<img src="/static/images/loader.svg" />
							<div>Checking your location</div>
						</div>

						<div class="current-location" v-if="currentLocation">
							<div class="label">Showing results based on your location</div>
							<div><span class="icon icon-location"></span>{{ this.currentLocation }}</div>
						</div>
					</transition>
				</div>
			</div>

			<!--Search results-->
			<div class="search-results" v-if="results">
				<!--No results found-->
				<transition name="fade">
					<div class="no-results" v-if="results && (!results.results || results.results.length === 0)">
						<h2>No results found</h2>
						<p>Try again with different keyword.</p>
					</div>
				</transition>

				<!--Query time-->
				<transition name="fade">
					<div class="query-time" v-if="results && results.results && results.results.length > 0">
						Took {{ results.took }} to search
					</div>
				</transition>

				<!--Search result items-->
				<transition-group name="list" tag="div" appear>
					<div class="search-item" v-for="(result, $index) in results.results" :key="result.id">
						<!--Top row-->
						<div class="search-item-header">
							<h2 class="name">
								<span>{{ result.fields.name }}</span>
								<span class="share icon icon-export" title="Share IFSC" @click="showSharePopup(result)"></span>
							</h2>
							<span class="ifsc hint--top" aria-label="Bank IFSC" v-clipboard="result.fields.IFSC" :key="$index"
								@success="handleClipboardEvent" @error="handleClipboardEvent">
								<label>IFSC</label>{{ result.fields.IFSC }}
							</span>
							<!--Hack to make IFSC selectable without overlapping with other content-->
							<span class="hack">&nbsp;</span>
						</div>

						<!--Second row-->
						<div class="search-item-body">
							<span class="branch hint--top" aria-label="Branch name"><label>Branch</label>{{ result.fields.branch }}</span>
							<div class="info">
								<div class="micr hint--top" v-if="result.fields.MICR" aria-label="MICR code">
									<label>MICR</label>
									<span class="icon icon-qrcode"></span>
									<span>{{ result.fields.MICR }}</span>
								</div>
								<div class="phone-number hint--top" v-if="result.fields.contact" aria-label="Bank contact no.">
									<label>Phone</label>
									<span class="icon icon-phone"></span>
									<span>{{ result.fields.contact.split(".")[0] }}</span>
								</div>
							</div>
						</div>

						<!--Last row-->
						<div class="search-item-footer">
							<div class="address hint--top" aria-label="Bank address">
								<label>Address</label>
								<span class="icon icon-location"></span><span>{{ result.fields.address }}</span>
							</div>
						</div>
					</div>
				</transition-group>
			</div>

			<transition name="fade">
				<div id="share-popup-wrapper" class="share-popup-wrapper" v-show="sharePopup && currentShareItem" @click="closeSharePopup">
					<div class="share-popup">
						<input class="share-url" :value="currentShareURL" readonly>
						<button class="btn hint--top" v-clipboard="currentShareURL" key="share" @success="handleClipboardEvent" @error="handleClipboardEvent"
							aria-label="Copy to clipboard">
							<span class="icon icon-clippy"></span>
						</button>
					</div>
				</div>
			</transition>
		</div>
	</div>
</template>

<script>
	import Vue from "vue"
	import axios from "axios"
	import smoothScroll from "smoothscroll"

	import ProgressBar from "./ProgressBar"

	export default Vue.extend({
		components: {
			"progress-bar": ProgressBar
		},
		data () {
			return {
				searchTerm: "",
				results: null,
				inputError: false,
				sharePopup: false,
				currentShareItem: {},
				isCheckingNavigation: false,
				currentLocation: "",
				shareBaseURI: process.env.SHARE_BASE
			}
		},
		mounted () {
			if (this.$route && this.$route.query && this.$route.query.q) {
				this.searchTerm = this.$route.query.q
				this.$nextTick(() => this.doSearch(this.searchTerm))
			} else {
				// Check for user location
				if (window.navigator && window.navigator.geolocation) {
					// Set navigation flag after a second since
					// when location access is denied the loader shows up
					// for fraction of time. This is not a ideal solution
					// since there can be a race condition where error callback is
					// triggered and after that the flag is set
					setTimeout(() => {
						this.isCheckingNavigation = true
					})

					window.navigator.geolocation.getCurrentPosition(this.geoSuccess, this.geoError)
				}
			}
		},
		watch: {
			searchTerm (val) {
				if (val.length > 2) {
					this.inputError = false
				}
			}
		},
		computed: {
			currentShareURL () {
				if (!this.currentShareItem || !this.currentShareItem.fields) return ""
				return process.env.SHARE_BASE + this.currentShareItem.fields.IFSC
			}
		},
		methods: {
			showSharePopup (item) {
				this.currentShareItem = item
				this.sharePopup = true
			},
			closeSharePopup (event) {
				if (event.target.id === "share-popup-wrapper") {
					this.currentShareItem = null
					this.sharePopup = false
				}
			},
			handleClipboardEvent (target, error) {
				let element = target.trigger
				let prevLabel = element.attributes && element.attributes["aria-label"] && element.attributes["aria-label"].value

				if (error) {
					element.attributes["aria-label"].value = "Copy failed."
				} else {
					element.attributes["aria-label"].value = "Copied."
				}

				if (prevLabel) {
					setTimeout(() => {
						element.attributes["aria-label"].value = prevLabel
					}, 1500)
				}
			},
			scrollToResults () {
				let width = Math.max(document.documentElement.clientWidth, window.innerWidth || 0)

				// Scroll to search for mobile devices.
				if (width <= 550) {
					smoothScroll(this.$el.querySelector(".search-results"))
				}
			},
			geoSuccess (position) {
				this.getUserLocation(position.coords.latitude, position.coords.longitude)
			},
			geoError () {
				this.isCheckingNavigation = false
			},
			startProgress () {
				this.$events.$emit("startProgress", true)
			},
			successProgress () {
				this.$events.$emit("finishProgress", {success: true})
			},
			errorProgress () {
				this.$events.$emit("finishProgress", {success: true})
			},
			doSearch (searchTerm, preventURIChange, preventCurrentLocationReset, preventScroll) {
				if (!searchTerm || searchTerm.length < 3) {
					this.inputError = true
					return
				}

				// Set current term as query param
				if (!preventURIChange) {
					this.$router.push({ path: "/", query: { q: searchTerm } })
				}

				// start progress bar
				this.startProgress()
				// Reset current loation
				if (!preventCurrentLocationReset) {
					this.currentLocation = ""
				}

				// prepare the params
				let params = {
					q: searchTerm
				}

				// API call
				axios.get(process.env.API_BASE + "/search", { params })
					.then((response) => {
						this.results = response.data
						this.successProgress()

						if (!preventScroll) {
							this.$nextTick(() => this.scrollToResults())
						}
					})
					.catch(() => {
						this.errorProgress()
					})
			},
			getFormattedAddressFromLocality (addresses) {
				let formatted = ""
				for (let address of addresses) {
					if (address.types.indexOf("sublocality") !== -1 || address.types.indexOf("locality") !== -1) {
						formatted += " " + address.long_name
					}
				}

				return formatted.toLowerCase()
			},
			getAddressFromLocation (addressResponse) {
				let locality
				let formatted
				let subLocality

				for (let i = addressResponse.length - 1; i >= 0; i--) {
					let address = addressResponse[i]

					if (address && address.types && (!locality || !subLocality)) {
						if (!locality && address.types.indexOf("locality") !== -1) {
							locality = address.formatted_address
							formatted = this.getFormattedAddressFromLocality(address.address_components)
						}

						if (!subLocality && address.types.indexOf("sublocality") !== -1) {
							subLocality = address.formatted_address
							formatted = this.getFormattedAddressFromLocality(address.address_components)
						}
					}
				}

				this.currentLocation = subLocality || locality

				// Do search based on the current formatted address
				if (formatted && formatted.length >= 3) {
					this.doSearch(formatted, true, true, true)
				}
			},
			getUserLocation (latitude, longitude) {
				let params = {
					latitude: latitude,
					longitude: longitude
				}

				axios.get(process.env.API_BASE + "/location", { params })
					.then((response) => {
						this.getAddressFromLocation(response.data.results)
						this.isCheckingNavigation = false
					})
					.catch(() => {
						this.isCheckingNavigation = false
					})
			}
		}
	})
</script>
