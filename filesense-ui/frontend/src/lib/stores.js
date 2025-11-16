import { writable } from 'svelte/store';

export const selectedPaths = writable(new Set());
