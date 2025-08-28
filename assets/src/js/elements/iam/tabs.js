import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

class CustomElement extends HTMLElement {
	template() {
		return html`
			<style>
				${css}

				:host {

				}
			</style>
			<c-link-tabs>
				<a href="/admin/iam/users"  >Users</a>
				<a href="/admin/iam/groups" >Groups</a>
				<a href="/admin/iam/roles"  >Roles</a>
			</c-link-tabs>
		`;
	}
	constructor() {
		super();

		this.attachShadow({mode:"open"});
	}

	async render() {
		render(template(this.params), this.shadowRoot);
	}

	get params() {
		return $.parsePathParam("/admin/iam")
	}
}
customElements.define("iam-tabs", CustomElement);
