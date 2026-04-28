import * as $ from "bm.js/bm.module.js";
import { html, render } from 'lit-html';

// c-input-tags
//
// Tag/chip input. Type a value, press Enter (or comma) to commit it as a
// chip. Backspace in an empty input removes the last chip. Click × on a
// chip to remove it.
//
// API:
//   .value : string[]              — get/set the current tag list
//   .suggestions : string[]        — set autocomplete options (datalist)
//   "change" event                 — fired with detail.tags after every change
//
// IME-safe: Enter is ignored while a composition (Korean, Japanese, etc.)
// is in progress.
class CustomElement extends HTMLElement {
    static get formAssociated() {
        return true;
    }

    #tags = [];
    #suggestions = [];
    #internal = this.attachInternals();

    get value() {
        return [...this.#tags];
    }
    set value(v) {
        this.#tags = v ? [...v] : [];
        this.#syncForm();
        this.render();
    }

    set suggestions(v) {
        this.#suggestions = v || [];
        this.render();
    }

    #syncForm() {
        this.#internal.setFormValue(this.#tags.join(","));
    }

    constructor() {
        super();
        this.attachShadow({ mode: "open" });
    }

    template() {
        return html`
            <style>
                :host {
                    display: inline-flex;
                    width: auto;
                    max-width: 100%;
                    box-sizing: border-box;
                    vertical-align: bottom;

                    font-family: system-ui, -apple-system, sans-serif;
                    background-color: white;
                    border: 1px solid #767676;
                    border-radius: 2px;
                    padding: 0px 4px;
                    cursor: text;
                    font-size: 0.8rem;
                }
                :host(:focus-within) {
                    border-color: black;
                    outline: 2px solid black;
                    outline-offset: -1px;
                }

                .tag-container {
                    display: flex;
                    flex-wrap: wrap;
                    align-items: center;
                    gap: 4px;
                    width: 100%;
                }

                .tag-chip {
                    display: inline-flex;
                    align-items: center;
                    background-color: #e0e0e0;
                    border: 1px solid #adadad;
                    border-radius: 3px;
                    padding: 0 4px;
                    font-size: 0.8rem;
                    user-select: none;
                }

                .tag-chip button {
                    display: inline-flex;
                    align-items: center;
                    justify-content: center;
                    background: none;
                    border: none;
                    color: #666;
                    margin-left: 4px;
                    cursor: pointer;
                    padding: 0;
                    font-size: 1rem;
                    line-height: 1;
                    width: 14px;
                    height: 14px;
                    border-radius: 50%;
                }
                .tag-chip button:hover {
                    background-color: #ccc;
                    color: black;
                }

                input {
                    border: none;
                    outline: none;
                    background: transparent;
                    flex-grow: 1;
                    font-size: inherit;
                    font-family: inherit;
                    padding: 2px 0;
                    margin: 0;
                    color: inherit;
                    field-sizing: content;
                }
            </style>
            <div class="tag-container" @click="${() => this.focusInput()}">
                ${this.#tags.map((tag, index) => html`
                    <span class="tag-chip">
                        ${tag}
                        <button type="button" @click="${e => this.removeTag(index, e)}">&times;</button>
                    </span>
                `)}

                <input type="text"
                    list="${this.id ? `${this.id}-suggestions` : "tag-suggestions"}"
                    placeholder="${this.attr('placeholder') || (this.#tags.length === 0 ? 'Tags...' : '')}"
                    @keydown="${evt => this.onKeyDown(evt)}"
                    @blur="${evt => this.onBlur(evt)}"
                />

                <datalist id="${this.id ? `${this.id}-suggestions` : 'tag-suggestions'}">
                    ${this.#suggestions.map(s => html`<option value="${s}"></option>`)}
                </datalist>
            </div>
        `;
    }

    focusInput() {
        this.shadowRoot.querySelector("input").focus();
    }

    #commit(inputEl) {
        const val = (inputEl.value || "").trim();
        inputEl.value = "";
        if (!val) return;
        if (this.#tags.includes(val)) return;
        this.#tags = [...this.#tags, val];
        this.#syncForm();
        this.dispatchChange();
        this.render();
    }

    onKeyDown(e) {
        // Korean / Japanese IME: ignore Enter while composing.
        if (e.isComposing) return;

        if (e.key === "Enter" || e.key === ",") {
            e.preventDefault();
            this.#commit(e.target);
            return;
        }
        if (e.key === "Backspace" && e.target.value === "" && this.#tags.length > 0) {
            e.preventDefault();
            this.#tags = this.#tags.slice(0, -1);
            this.#syncForm();
            this.dispatchChange();
            this.render();
        }
    }

    onBlur(e) {
        // Auto-commit on blur so a typed-but-not-Entered value isn't lost.
        this.#commit(e.target);
    }

    removeTag(index, e) {
        e.stopPropagation();
        this.#tags = this.#tags.filter((_, i) => i !== index);
        this.#syncForm();
        this.dispatchChange();
        this.render();
    }

    dispatchChange() {
        this.fireEvent("change", { tags: this.value });
    }

    connectedCallback() {
        this.render();
    }

    render() {
        render(this.template(), this.shadowRoot);
        this.attr("count", this.#tags.length);
    }
}

customElements.define("c-input-tags", CustomElement);
