import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';
import {css} from "common.js";

class CustomElement extends HTMLElement {
	template() {
		return html`
			<style>
				${css}

				:host {
				}
				table {
					th, td {
						border: 1px solid var(--gray-400);
					}
				}
			</style>
			<table>
				<thead>
					<tr>
						<th>Name</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
				${this.#items.map((item) => html`
					<tr>
						<td>${item.name}</td>
						<td>
							<a href="#">edit</a>
						</td>
					</tr>
				`)}
				</tbody>
			</table>
		`;
	}
	constructor() {
		super();

		this.attachShadow({mode: "open"});
	}

	#items = [];

	async render() {
		render(tmpl(this), this.shadowRoot);
	}

	async onConnected() {
		let res = await $.request("GET", "/api/v1/roles");

		this.#items = res.json;

		this.render();
	}
}
customElements.define("iam-roles-list", CustomElement);
