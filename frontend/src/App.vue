<template>
	<div id="app" class="full-height">
		<div class="main">
			<!--Header-->
			<div id="header"></div>
			<!--Search component-->
			<search></search>
			<!--Footer-->
			<div class="footer">
				<div class="links">
					<a href="https://github.com/vividvilla/bankr/wiki/API" target="_blank">Api</a> &mdash;
					<a href="https://github.com/vividvilla/bankr" target="_blank">Source</a> &mdash;
					<a href="https://goo.gl/forms/0tioEeHq2KtGW6lT2" target="_blank">
						<span class="icon icon-mail"></span>  bugs/feedback/contact
					</a>
				</div>
				<div class="easter-egg">Proprietor Vishnu</div>
			</div>
		</div>
	</div>
</template>

<script>
import Search from "./components/Search"

export default {
	name: "app",
	components: {
		Search
	},
	mounted () {
	},
	methods: {
	}
}
</script>

<style src="./assets/styles/app.scss" lang="scss"></style>
