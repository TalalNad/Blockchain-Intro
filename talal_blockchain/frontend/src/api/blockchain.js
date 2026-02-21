import { api } from "./client";

export async function getChain() {
  const { data } = await api.get("/chain");
  return data;
}

export async function getPending() {
  const { data } = await api.get("/pending");
  return data;
}

export async function addTx(tx) {
  const { data } = await api.post("/tx", { tx });
  return data;
}

export async function mine() {
  const { data } = await api.post("/mine");
  return data;
}

export async function searchTx(q) {
  const { data } = await api.get(`/search?q=${encodeURIComponent(q)}`);
  return data;
}