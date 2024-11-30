<template>
  <div id="app">
    <header>
      <h1>Validate your pip requirements.txt in seconds!</h1>
    </header>
    <main>
      <FileDrop @file-submitted="handleFileSubmission" />
      <h4>How does it work?</h4>
      <p>The program will parse out your requirements file and access APIs to ensure the versions are good.
      If your file's versions are verified, the program will pip install your requirements file in a separate
      environment, ensuring that the requirements are all packaged together in your file. This process will take
      a few minutes depending on how large your requirements file is.</p>
      <div class="output-container">
        <div v-if="loading" class="loading-spinner"></div>
        <div v-else-if="output" class="output">
          {{ output }}
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import FileDrop from "./components/FileDrop.vue";

export default {
  components: {
    FileDrop,
  },
  data() {
    return {
      loading: false,
      output: null,
    };
  },
  methods: {
    async handleFileSubmission(file) {
      this.loading = true;
      this.output = null;

      try {
        const formData = new FormData();
        formData.append("file", file);
        
        // Need to edit this when I get to it...
        const response = await fetch("http://localhost:8080", {
          method: "POST",
          body: formData,
        });

        if (response.ok) {
          const jsonResponse = await response.json();
          this.output = jsonResponse.prettyOutput;
          this.details = jsonResponse.details;
        } else {
          this.output = "Error validating the file. Please try again.";
        }
      } catch (error) {
        this.output = `An error occurred: ${error.message}`;
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>

<style>
body {
  margin: 0;
  font-family: Arial, sans-serif;
  background-color: #121212;
  color: #e0e0e0;
}

#app {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  min-height: 100vh;
  padding: 1rem;
}

header h1 {
  margin: 1rem 0;
  font-size: 2rem;
}

main {
  width: 100%;
  max-width: 600px;
}

.output-container {
  margin-top: 2rem;
  min-height: 150px;
}

.loading-spinner {
  border: 6px solid #e0e0e0;
  border-top: 6px solid #6200ea;
  border-radius: 50%;
  width: 50px;
  height: 50px;
  margin: 0 auto;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.output {
  padding: 1rem;
  background-color: #1e1e1e;
  border: 1px solid #333;
  border-radius: 8px;
  word-wrap: break-word;
}
</style>
