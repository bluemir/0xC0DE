import {html, render} from 'lit-html';
import {css, errorMessage} from "@/common.js";
import * as passkey from "@/passkey.js";

var tmpl = (elem) => html`
	<style>
		${css}

		table {
			th, td {
				border: 1px solid var(--gray-400);
			}
		}
		p[role="alert"] {
			color: var(--red-600);
		}
	</style>
	<form @submit=${(evt) => elem.add(evt)}>
		<input name="label" placeholder="passkey 이름" />
		<button ?disabled=${elem.busy || !passkey.supported()}>Add Passkey</button>
	</form>
	${elem.message ? html`<p role="alert">${elem.message}</p>` : ""}
	<table>
		<thead>
			<tr>
				<th>label</th>
				<th>created</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
		${elem.items.map((item) => html`
			<tr>
				<td>${item.label || "-"}</td>
				<td>${new Date(item.createdAt).toLocaleString()}</td>
				<td><button @click=${() => elem.revoke(item.index)}>Delete</button></td>
			</tr>
		`)}
		</tbody>
	</table>
`;

// <passkey-list> 는 계정에 등록된 passkey 를 관리한다.
class CustomElement extends HTMLElement {
	constructor() {
		super();

		this.attachShadow({mode: "open"});

		this.items = [];
		this.busy = false;
		this.message = "";
	}

	async render() {
		render(tmpl(this), this.shadowRoot);
	}

	async onConnected() {
		await this.reload();
	}

	async reload() {
		try {
			this.items = await passkey.list();
		} catch(e) {
			console.error(e);

			this.message = errorMessage(e);
		}
		this.render();
	}

	async add(evt) {
		evt.preventDefault();

		this.busy = true;
		this.message = "";
		this.render();

		try {
			let label = new FormData(evt.target).get("label");

			await passkey.register({label});

			evt.target.reset();
		} catch(e) {
			console.error(e);

			this.message = errorMessage(e);
		}

		this.busy = false;
		await this.reload();
	}

	async revoke(index) {
		this.message = "";

		try {
			await passkey.revoke(index);
		} catch(e) {
			console.error(e);

			this.message = errorMessage(e);
		}

		await this.reload();
	}
}
customElements.define("passkey-list", CustomElement);
