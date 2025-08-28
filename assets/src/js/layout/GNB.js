import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';
import {css} from "common.js";

class GlobalNavigationBar extends HTMLElement {
	template() {
		return html`
			<style>
				${css}

				:host {
					display: flex;
					justify-content: space-between;
				}
				a {
					color: inherit;
					text-decoration: none;
				}
				a:hover {
					text-decoration: underline;
				}
			</style>
			<section id="logo">
				<a href="/">0xC0DE</a>
				${this.hasAttribute("admin") ? html`
					 <a href="/admin">Admin</a>
				 `:""}
			</section>
			<section id="action">
				${this.user ? html`
					<a href="/admin"><c-icon kind="construction" /></a>
					<a href="/users/profile">Profile</a>
					<a href="/users/logout" >Logout</a>
				` : html`
					<a href="/users/login">Login</a>
					<a href="/users/register">Register</a>
				`}
			</section>
		`;
	}
	constructor() {
		super();

		this.attachShadow({mode: 'open'})
	}

	async render() {
		render(this.template(), this.shadowRoot);
	}

	async onConnected() {
		let res = await $.request("GET", `/api/v1/users/me`);

		this.user = res.json;

		this.render();
	}
}
customElements.define("global-navigation-bar", GlobalNavigationBar);

