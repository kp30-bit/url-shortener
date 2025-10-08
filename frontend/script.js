const shortenBtn = document.getElementById("shortenBtn");
const copyBtn = document.getElementById("copyBtn");
const deleteBtn = document.getElementById("deleteBtn");
const urlInput = document.getElementById("urlInput");
const shortURL = document.getElementById("shortURL");
const resultDiv = document.getElementById("result");

const totalURLsSpan = document.getElementById("totalURLs");
const totalClicksSpan = document.getElementById("totalClicks");

let currentShortId = null;

// -------------------------
// Fetch Analytics
// -------------------------
async function fetchAnalytics() {
  try {
    const res = await fetch("/analytics");
    if (!res.ok) throw new Error("Failed to fetch analytics");

    const data = await res.json();
    totalURLsSpan.textContent = data.total_urls;
    totalClicksSpan.textContent = data.total_clicks;
  } catch (err) {
    console.error("Analytics error:", err);
  }
}

// Call analytics on page load
window.addEventListener("DOMContentLoaded", fetchAnalytics);

// -------------------------
// Shorten URL
// -------------------------
shortenBtn.addEventListener("click", async () => {
  let longURL = urlInput.value.trim();
  if (!longURL) {
    alert("Please enter a valid URL 😅");
    return;
  }

  // Ensure URL starts with http:// or https://
  if (!/^https?:\/\//i.test(longURL)) {
    longURL = "http://" + longURL;
  }

  try {
    const res = await fetch("/shorten", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ original_url: longURL }), // match backend field
    });

    if (!res.ok) throw new Error("Failed to shorten URL");

    const data = await res.json();

    // Display short URL without API prefix
    const shortened = `${window.location.origin}/${data.short_id}`;

    currentShortId = data.short_id;
    shortURL.value = shortened;
    resultDiv.classList.remove("hidden");

    // Refresh analytics after shortening
    fetchAnalytics();
  } catch (err) {
    console.error(err);
    alert("Something went wrong! Please try again 😢");
  }
});

// -------------------------
// Copy short URL
// -------------------------
copyBtn.addEventListener("click", () => {
  navigator.clipboard.writeText(shortURL.value);
  copyBtn.textContent = "✅ Copied!";
  setTimeout(() => (copyBtn.textContent = "📋 Copy"), 1500);
});

// -------------------------
// Delete URL
// -------------------------
deleteBtn.addEventListener("click", async () => {
  if (!currentShortId) return alert("No URL to delete!");

  if (!confirm("Are you sure you want to delete this URL?")) return;

  try {
    const res = await fetch(`/${currentShortId}`, {
      method: "DELETE",
    });
    if (!res.ok) throw new Error("Failed to delete URL");

    alert("🗑️ URL deleted successfully!");
    resultDiv.classList.add("hidden");
    urlInput.value = "";

    // Refresh analytics after deletion
    fetchAnalytics();
  } catch (err) {
    console.error(err);
    alert("Failed to delete the link!");
  }
});
