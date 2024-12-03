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
        a few minutes depending on how large your requirements file is.
      </p>
      <div class="output-container">
        <div v-if="loading" class="loading-spinner"></div>
        <div v-else>
          <div v-if="output" class="output">
            <h4>API Check Output:</h4>
            <div v-html="formatOutput(output)"></div>
          </div>
          <div v-if="dockerDetails" class="output">
            <h4>Environment Install Details:</h4>
            <div v-html="formatOutput(dockerDetails)"></div>
          </div>
          <p>Note: This will fail if not all packages are verified!</p>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import FileDrop from "./components/FileDrop.vue";
import secret from '@aws-amplify/backend'

export default {
  components: {
    FileDrop,
  },
  data() {
    return {
      loading: false,
      output: null,
      dockerDetails: null,
    };
  },
  methods: {
    async handleFileSubmission(file) {
      this.loading = true;
      this.output = null;
      this.dockerDetails = null;

      try {
        const formData = new FormData();
        formData.append("file", file);
        let url = `${import.meta.env.VITE_API_URL}`
        if (url == "") {
            url = secret('VITE_API_URL')
        }
        console.log(url)
        let auth = `${import.meta.env.VITE_AUTH_TOKEN}`
        if (auth == "") {
            auth = secret('VITE_AUTH_TOKEN')
        }
        const response = await fetch(url, {
            method: "POST",
            headers: {
              Authorization: `Bearer ` + auth,
            },
            body: formData,
        });

        if (response.ok) {
          const jsonResponse = await response.json();
          this.output = jsonResponse.prettyOutput;
          this.dockerDetails = jsonResponse.installOutput;
        } else {
          this.output = "Error validating the file. Please try again.";
        }
      } catch (error) {
        this.output = `An error occurred: ${error.message}`;
      } finally {
        this.loading = false;
      }
    },
    formatOutput(text) {
      return text.replace(/\n/g, "<br>");
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
  margin-bottom: 1rem;
}
</style>
