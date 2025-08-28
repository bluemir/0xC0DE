import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

var tmpl = (elem) => html`
	<style>
		${css}

		:host {
		}
		a {
			display: block;
			text-decoration: none;
			padding: 0.5rem 0;
		}
		a:hover {
			background: var(--gray-100);
		}
		a.selected {
			background: var(--green-100);
		}
	</style>
	<a href="/admin/users" >IAM</a>
	<a href="/admin/server">Server</a>
	<a href="/admin/events">Events</a>
`;

class AdminMenu extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({mode: "open"});
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}

	onConnected() {
		$.all(this.shadowRoot, "a").
			filter((elem, index) => elem.hasAttribute("exact") ? elem.attr("href") == location.pathname: location.pathname.startsWith(elem.attr("href"))).
			forEach(elem => elem.classList.add("selected"));
	}
}
customElements.define("admin-menu", AdminMenu);
