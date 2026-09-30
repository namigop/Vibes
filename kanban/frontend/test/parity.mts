// Cross-layer check: runs the frontend's real ordering helpers against a live
// API, to prove the index the UI computes is the index the server settles on.
//
//   npm run parity            # uses NEXT_PUBLIC_API_BASE_URL or the default below
//   KANBAN_API=... npm run parity
//
// The Go server must already be running. Run it with `make api`.

import { createBoardState, moveCardInState } from "../src/lib/board-state.ts";

const API = process.env.KANBAN_API ?? "http://localhost:8081/api/v1";
let failures = 0;

function check(name, cond, detail = "") {
  if (cond) {
    console.log(`  PASS  ${name}`);
  } else {
    console.log(`  FAIL  ${name}  ${detail}`);
    failures++;
  }
}

async function api(path, init) {
  const res = await fetch(`${API}${path}`, init);
  if (res.status === 204) return null;
  return res.json();
}

// A scratch board so the seeded boards are untouched.
const board = await api("/boards", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ name: "Parity Check" }),
});
const detail = await api(`/boards/${board.id}`);

const columns = detail.columns;
const todo = columns.find((c) => c.name === "To Do");
const review = columns.find((c) => c.name === "Review");

for (const title of ["one", "two", "three"]) {
  await api(`/columns/${todo.id}/cards`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title }),
  });
}
for (const title of ["alpha", "beta"]) {
  await api(`/columns/${review.id}/cards`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title }),
  });
}

const fresh = await api(`/boards/${board.id}`);
let state = createBoardState(fresh.board, fresh.columns, fresh.cards);

console.log("\n=== Frontend/server move parity ===");

// Cross-column: drag "one" (index 0 of To Do) to index 1 in Review.
const movingId = state.cardsByColumn[todo.id][0].id;
const optimistic = moveCardInState(state, movingId, review.id, 1);
const predicted = optimistic.cardsByColumn[review.id].map((c) => c.title);

const saved = await api(`/cards/${movingId}/move`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ column_id: review.id, position: 1 }),
});
check("server accepted the move", saved && saved.column_id === review.id, JSON.stringify(saved));
check("server settled on the index the UI predicted", saved.position === 1, `server=${saved?.position} ui=${optimistic.cardsByColumn[review.id].find((c) => c.id === movingId).position}`);

const after = await api(`/boards/${board.id}`);
state = createBoardState(after.board, after.columns, after.cards);
const actual = state.cardsByColumn[review.id].map((c) => c.title);
check(
  "rendered order matches the optimistic prediction exactly",
  JSON.stringify(actual) === JSON.stringify(predicted),
  `server=${JSON.stringify(actual)} ui=${JSON.stringify(predicted)}`,
);

const todoPositions = state.cardsByColumn[todo.id].map((c) => c.position);
check("source column reindexed gap-free", JSON.stringify(todoPositions) === "[0,1]", JSON.stringify(todoPositions));
const reviewPositions = state.cardsByColumn[review.id].map((c) => c.position);
check("dest column reindexed gap-free", JSON.stringify(reviewPositions) === "[0,1,2]", JSON.stringify(reviewPositions));

// Within-column: drag the last card of Review to the front.
const lastId = state.cardsByColumn[review.id][2].id;
const withinPrediction = moveCardInState(state, lastId, review.id, 0);
const withinPredicted = withinPrediction.cardsByColumn[review.id].map((c) => c.title);
const saved2 = await api(`/cards/${lastId}/move`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ column_id: review.id, position: 0 }),
});
check("within-column move settled at 0", saved2.position === 0, `got ${saved2?.position}`);

const after2 = await api(`/boards/${board.id}`);
const state2 = createBoardState(after2.board, after2.columns, after2.cards);
check(
  "within-column order matches prediction",
  JSON.stringify(state2.cardsByColumn[review.id].map((c) => c.title)) === JSON.stringify(withinPredicted),
  `server=${JSON.stringify(state2.cardsByColumn[review.id].map((c) => c.title))} ui=${JSON.stringify(withinPredicted)}`,
);

// Clean up.
await api(`/boards/${board.id}`, { method: "DELETE" });
check("scratch board deleted", true);

console.log(`\nfailures: ${failures}`);
process.exit(failures === 0 ? 0 : 1);
