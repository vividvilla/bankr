<template>
	<div id="app">
		<div class="main">
			<!--Header-->
			<div id="header"></div>
			<!--Search component-->
			<search></search>
			<!--Footer-->
			<div class="footer"></div>
		</div>
	</div>
</template>

<script>
import Search from "./components/Search"

export default {
	name: "app",
	components: {
		Search
	}
}
</script>

<style src="./assets/styles/app.scss" lang="scss"></style>
