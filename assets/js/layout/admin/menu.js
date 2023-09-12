import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';

var tmpl = (app) => html`
	<style>
		@import "/static/css/root.css";

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
	<a part="item" href="/admin/users" >Users</a>
	<a part="item" href="/admin/groups">Group</a>
	<a part="item" href="/admin/roles" >Role</a>
	<a part="item" href="/admin/server">Server</a>
	<a part="item" href="/admin/events">Events</a>
`;

class AdminMenu extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadow);
	}

	onConnected() {
		console.log(location.pathname)
		$.all(this.shadowRoot, "a").filter( elem => elem.attr("href") == location.pathname).forEach(elem => elem.classList.add("selected"));
	}
}
customElements.define("admin-menu", AdminMenu);
