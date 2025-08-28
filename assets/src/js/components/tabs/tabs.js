import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

/*
<c-tabs selected="a">
	<label   slot="header" tab="a">A Header</label>
	<section slot="panel"  tab="a">
		A Contents
	</section>
	<label   slot="header" tab="b">B Header</label>
	<section slot="panel"  tab="b">
		B Contents
	</section>
</c-tabs>
*/

var tmpl = (app) => html`
	<style>
		${css}

		:host {
		}
		header {
			display: flex;
			gap: 1rem;

			margin-bottom: 1rem;
		}
		::slotted([slot=header]) {
			padding: 0.5rem 1rem;
			border-bottom: 0.2rem solid var(--gray-200);
		}
		::slotted([slot=header].selected) {
			border-bottom: 0.2rem solid var(--green-400);
		}
		::slotted([slot=panel]) {
			display: none;
		}
		::slotted([slot=panel].selected) {
			display: block;
		}
	</style>
	<header @click=${evt => app.handleTabHeaderClick(evt)}>
		<slot name="header"></slot>
	</header>
	<slot name="panel"></slot>
`;

class Tabs extends HTMLElement {
	constructor() {
		super();
		this.attachShadow({mode: "open"});
	}
	static get observedAttributes() {
		return [ "selected" ];
	}
	async onConnected(){
		let tab = this.attr("selected") || $.get(this, `c-tab-header`).attr("tab");

		this.changePanel(tab);
	}
	async onAttributeChanged(name, ov, nv) {
		switch(name) {
			case "selected":
				if (ov == nv) {
					return;
				}
				this.changePanel(nv);
				return
		}
	}
	async render() {
		render(tmpl(this), this.shadowRoot);
	}
	async handleTabHeaderClick(evt) {
		let tab = evt.target.attr("tab");
		if (!tab) {
			return
		}
		this.selected = tab;
	}

	get selected() {
		return this.attr("selected");
	}
	set selected(tab) {
		this.attr("selected", tab);
	}
	changePanel(tab) {
		$.all(this, `[slot=header]`).forEach(elem => elem.classList.remove("selected"));
		$.get(this, `[slot=header][tab=${tab}]`).classList.add("selected");
		$.all(this, `[slot=panel]`).forEach(elem => elem.classList.remove("selected"));
		$.get(this, `[slot=panel][tab=${tab}]`).classList.add("selected");

		$.all(this, `.selected`).forEach(e => {
			e.fireEvent("active")
		});
	}
}
customElements.define("c-tabs", Tabs);
