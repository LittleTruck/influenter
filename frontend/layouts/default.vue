<template>
  <div class="layout-default">
    <header class="header">
      <div class="container">
        <h1 class="logo">Influenter</h1>
        <nav class="nav">
          <!-- 導航項目將在後續添加 -->
        </nav>
      </div>
    </header>
    
    <main class="main">
      <div class="container">
        <slot />
      </div>
    </main>
    
    <footer class="footer">
      <div class="container">
        <p>&copy; 2025 Influenter. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
// Default Layout
</script>

<style scoped>
.layout-default {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background-color: #2c3e50;
  color: white;
  padding: 1rem 0;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

.header .container {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  font-size: 1.5rem;
  font-weight: bold;
  margin: 0;
}

.nav {
  display: flex;
  gap: 1rem;
}

.main {
  flex: 1;
  padding: 2rem 0;
}

.footer {
  background-color: #34495e;
  color: white;
  padding: 1.5rem 0;
  text-align: center;
  margin-top: auto;
}

.footer p {
  margin: 0;
}
</style>


