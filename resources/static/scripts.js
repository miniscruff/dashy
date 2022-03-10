window.addEventListener('DOMContentLoaded', async () => {
  const response = await fetch('/api/values');
  const values = await response.json();
	console.log(values);
});
