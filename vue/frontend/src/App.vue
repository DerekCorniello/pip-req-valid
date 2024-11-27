<template>
  <div id="app">
    <header>
      <h1>Validate your pip requirements.txt in seconds!</h1>
    </header>
    <main>
      <FileDrop @file-submitted="handleFileSubmission" />
      <div class="options">
        <label>
          <input type="checkbox" v-model="runDockerInstall" />
          Run Docker installation validation (this may take extra time)
        </label>
      </div>
      <div class="output-container">
        <div v-if="loading" class="loading-spinner"></div>
        <div v-else-if="output" class="output">
          <div>
            <h2>Validation Results:</h2>
            <pre>{{ output }}</pre>
            <div v-if="details">
              <h3>Additional Details:</h3>
              <pre>{{ details }}</pre>
            </div>
          </div>
        </div>
        <div v-if="dockerInfo">
          <p><strong>Note:</strong> Docker validation install is in progress. This will take extra time.</p>
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
      details: null,
      runDockerInstall: false, // flag for Docker installation
      dockerInfo: false, // show additional info about Docker processing
    };
  },
  methods: {
    async handleFileSubmission(file) {
      this.loading = true;
      this.output = null;
      this.dockerInfo = false;
      this.details = null;

      try {
        const formData = new FormData();
        formData.append("file", file);
        formData.append("runDockerInstall", this.runDockerInstall); // Send the flag

        const response = await fetch("backendservice", {
          method: "POST",
          body: formData,
        });

        if (response.ok) {
          const jsonResponse = await response.json();
          this.output = jsonResponse.prettyOutput;
          this.details = jsonResponse.details;

          // If Docker install is requested, show progress message
          if (this.runDockerInstall) {
            this.dockerInfo = true;
          }
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

<style scoped>
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

.options {
  margin-top: 20px;
  font-size: 14px;
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
