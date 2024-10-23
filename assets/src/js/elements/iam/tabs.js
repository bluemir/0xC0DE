import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

var tmpl = (elem, param) => html`
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

class CustomElement extends $.CustomElement {
	constructor() {
		super();
	}

	async render() {
		render(tmpl(this, this.params), this.shadowRoot);
	}

	get params() {
		return $.parsePathParam("/admin/iam")
	}
}
customElements.define("iam-tabs", CustomElement);
