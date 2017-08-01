import * as URI from "urijs";
import * as querystring from "querystring";

export interface SearchParams {
	query: string;
	repos: string;
	files: string;
	matchCase: boolean;
	matchWord: boolean;
	matchRegex: boolean;
}

export function handleSearchInput(e: any, readSearchParamsFromURL: boolean): void {
	const query = e.target.value;
	if ((e.key !== "Enter" && e.keyCode !== 13) || !query) {
		return;
	}

	const params = readSearchParamsFromURL ? getSearchParamsFromURL(window.location.href) : getSearchParamsFromLocalStorage();
	params.query = query;

	let newTab = false;
	if (e.metaKey || e.altKey || e.ctrlKey) {
		newTab = true;
	}
	const path = getSearchPath(params);
	newTab ? window.open(path, "_blank") : window.location.href = path;
}

export function getSearchPath(params: SearchParams): string {
	return `/search?q=${encodeURIComponent(params.query)}&repos=${encodeURIComponent(params.repos)}${params.files ? `&files=${encodeURIComponent(params.files)}` : ""}${params.matchCase ? "&matchCase=true" : ""}${params.matchWord ? "&matchWord=true" : ""}${params.matchRegex ? "&matchRegex=true" : ""}`;
}

export function getSearchParamsFromURL(url: string): SearchParams {
	const query: { [key: string]: string } = querystring.parse(URI.parse(url).query);
	return {
		query: query["q"] || "",
		repos: query["repos"] || "active",
		files: query["files"] || "",
		matchCase: query["matchCase"] === "true",
		matchWord: query["matchWord"] === "true",
		matchRegex: query["matchRegex"] === "true",
	}
}

export function getSearchParamsFromLocalStorage(): SearchParams {
	return {
		query: window.localStorage.getItem("searchQuery") || "",
		repos: window.localStorage.getItem("searchRepoScope") || "active",
		files: window.localStorage.getItem("searchFileScope") || "",
		matchCase: window.localStorage.getItem("searchMatchCase") === "true",
		matchWord: window.localStorage.getItem("searchMatchWord") === "true",
		matchRegex: window.localStorage.getItem("searchMatchRegex") === "true",
	}
}
