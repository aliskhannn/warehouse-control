const apiBase = "http://localhost:8080/api";
let role = "viewer";
let token = "";

document.getElementById("loginBtn").onclick = loginUser;
document.getElementById("registerBtn").onclick = registerUser;

async function registerUser() {
  const username = document.getElementById("username").value;
  const password = document.getElementById("password").value;
  const selectedRole = document.getElementById("role").value;

  try {
    const res = await fetch(`${apiBase}/auth/register`, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({ username, password, role: selectedRole })
    });
    const data = await res.json();
    if (!res.ok) {
      document.getElementById("loginMsg").innerText = data.error || "Registration failed";
      return;
    }
    document.getElementById("loginMsg").innerText = "Registered successfully, now login.";
  } catch (err) {
    document.getElementById("loginMsg").innerText = "Error: " + err.message;
  }
}

async function loginUser() {
  const username = document.getElementById("username").value;
  const password = document.getElementById("password").value;
  role = document.getElementById("role").value;

  try {
    const res = await fetch(`${apiBase}/auth/login`, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({ username, password })
    });
    const data = await res.json();
    if (!res.ok) {
      document.getElementById("loginMsg").innerText = data.error || "Login failed";
      return;
    }
    token = data.result.token;

    document.querySelector(".login").style.display = "none";
    document.querySelector(".app").style.display = "block";

    if (role === "admin" || role === "manager") {
      document.getElementById("addItemBtn").style.display = "inline-block";
    }
    loadItems();
  } catch (err) {
    document.getElementById("loginMsg").innerText = "Error: " + err.message;
  }
}

async function loadItems() {
  try {
    const res = await fetch(`${apiBase}/items`, { headers: { Authorization: `Bearer ${token}` } });
    const data = await res.json();
    if (!res.ok) { alert(data.error || "Failed to load items"); return; }

    const items = data.result;
    const tbody = document.getElementById("itemsTable");
    tbody.innerHTML = "";
    items.forEach(item => {
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td>${item.id}</td>
        <td>${item.name}</td>
        <td>${item.description}</td>
        <td>${item.quantity}</td>
        <td>${item.price}</td>
        <td>
          ${(role === "admin" || role === "manager") ? `<button onclick="editItem('${item.id}')">Edit</button>` : ""}
          ${(role === "admin") ? `<button onclick="deleteItem('${item.id}')">Delete</button>` : ""}
          ${(role === "admin") ? `<button onclick="showHistory('${item.id}')">History</button>` : ""}
        </td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {
    alert("Failed to load items: " + err.message);
  }
}

async function editItem(id) {
  const name = prompt("Name:");
  const description = prompt("Description:");
  const quantity = parseInt(prompt("Quantity:"));
  const price = parseFloat(prompt("Price:"));

  try {
    const res = await fetch(`${apiBase}/items/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
      body: JSON.stringify({ name, description, quantity, price })
    });
    const data = await res.json();
    if (!res.ok) { alert(data.error || "Failed to update"); return; }
    loadItems();
  } catch (err) { alert("Error: " + err.message); }
}

async function deleteItem(id) {
  if (!confirm("Delete this item?")) return;

  try {
    const res = await fetch(`${apiBase}/items/${id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` }
    });
    const data = await res.json();
    if (!res.ok) { alert(data.error || "Failed to delete"); return; }
    loadItems();
  } catch (err) { alert("Error: " + err.message); }
}

async function showHistory(id) {
  try {
    const res = await fetch(`${apiBase}/audit/items/${id}/history`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    const data = await res.json();
    if (!res.ok) { alert(data.error || "Failed to fetch history"); return; }

    const history = data.result;
    const tbody = document.getElementById("historyTable");
    tbody.innerHTML = "";
    history.forEach(h => {
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td>${h.action}</td>
        <td>${h.changed_by}</td>
        <td>${h.changed_at}</td>
        <td><pre>${JSON.stringify(h.old_data, null, 2)}</pre></td>
        <td><pre>${JSON.stringify(h.new_data, null, 2)}</pre></td>
      `;
      tbody.appendChild(tr);
    });

    document.getElementById("historyModal").style.display = "block";
  } catch (err) { alert("Error: " + err.message); }
}

document.querySelector(".close").onclick = () => {
  document.getElementById("historyModal").style.display = "none";
};

document.getElementById("addItemBtn").onclick = async () => {
  const name = prompt("Name:");
  const description = prompt("Description:");
  const quantity = parseInt(prompt("Quantity:"));
  const price = parseFloat(prompt("Price:"));

  try {
    const res = await fetch(`${apiBase}/items`, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
      body: JSON.stringify({ name, description, quantity, price })
    });
    const data = await res.json();
    if (!res.ok) { alert(data.error || "Failed to add"); return; }
    loadItems();
  } catch (err) { alert("Error: " + err.message); }
};
