import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';

var tmpl = (elem) => html`
	<style>
		:host {
			display: flex;
			justify-content: space-between;
		}
		a {
			color: var(--fg-color, black);
			text-decoration: none;
		}
		a:hover {
			text-decoration: underline;
		}
	</style>
	<section id="logo">
		<a href="/">LOGO</a>
		${elem.hasAttribute("admin") ? html`
			 <a href="/admin">Admin</a>
		 `:""}
	</section>
	<section id="action">
		${ elem.user ? html`
			<a href="/admin"><c-icon kind="construction" /></a>
			<a href="/users/profile">Profile</a>
			<a href="/users/logout" >Logout</a>
		` : html`
			<a href="/users/login">Login</a>
			<a href="/users/register">Register</a>
		`}
	</section>
`;

class GlobalNavigationBar extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this), this.shadow);
	}

	async onConnected() {
		let res = await $.request("GET", `/api/v1/users/me`);

		this.user = res.json;

		this.render();
	}
}
customElements.define("global-navigation-bar", GlobalNavigationBar);

