import * as $ from "bm.js/bm.module.js";
//import {html, render} from '/lib/lit-html/lit-html.js';
import {html, render} from 'lit-html';

var tmpl = (elem) => html`
	<style>
		@import "/static/css/root.css";

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
		${elem.items.map((item) => html`
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

class CustomElement extends $.CustomElement {
	constructor() {
		super();

		this.items = [];
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}

	async onConnected() {
		let res = await $.request("GET", "/api/v1/roles");

		this.items = res.json;

		this.render();
	}
}
customElements.define("iam-roles-list", CustomElement);
