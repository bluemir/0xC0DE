import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';

var tmpl = (app) => html`
	<style>
		@import "/static/css/root.css";

		:host {
			position: fixed;
			width: 100%;
			height: 100%;

			display: grid;
			grid-template-columns: auto 1fr;
			grid-template-rows: auto 1fr auto;
		}

		header, main, footer{
			padding: 0 1rem;
		}
		aside, main {
			overflow-y: scroll;
		}

		header {
			grid-column: span 2;
			background: var(--green-800);
		}
		aside {
			min-width: 15rem;
			grid-row: span 2;
		}
		main {
		}
		footer {
		}

		global-navigation-bar {
			--fg-color: white;
			margin: 0.3rem 0;
		}
		admin-menu::part(item) {
			padding: 0.5rem 1rem;
		}

	</style>
	<header>
		<global-navigation-bar admin></global-navigation-bar>
	</header>
	<aside>
		<admin-menu></admin-menu>
	</aside>
	<main>
		<slot></slot>
	</main>
	<footer>
		<admin-footer></admin-footer>
	</footer>
`;

class AdminPage extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}
}
customElements.define("admin-page", AdminPage);
