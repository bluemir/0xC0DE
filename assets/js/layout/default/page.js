import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';

var tmpl = (elem) => html`
	<style>
		@import "/static/css/root.css";

		:host {
		}

		global-navigation-bar, ::slotted(*) {
			max-width: 1280px;
			margin: auto;
		}
		global-navigation-bar {
			--fg-color: white;
			padding: 0.3rem 1rem;
		}
		::slotted(*) {
			padding: 0 1rem;
		}

		header {
			display: fixed;
			background: var(--blue-alt-700);
		}

		main {
			padding-top: 1rem;
		}
	</style>
	<header>
		<global-navigation-bar></global-navigation-bar>
	</header>
	<main>
		<slot></slot>
	</main>
`;

class DefaultPage  extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadow);
	}
}
customElements.define("default-page", DefaultPage);
