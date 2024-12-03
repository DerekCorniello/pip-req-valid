<template>
  <div id="app">
    <header>
      <h1>Validate your pip requirements.txt in seconds!</h1>
    </header>
    <main>
      <FileDrop @file-submitted="handleFileSubmission" />
      <h4>How does it work?</h4>
      <p>
        The program will parse out your requirements file and access APIs to ensure the versions are good.
        If your file's versions are verified, the program will pip install your requirements file in a separate
        environment, ensuring that the requirements are all packaged together in your file. This process will take
        a few minutes depending on how large your requirements file is.
      </p>
      <div class="output-container">
        <div v-if="loading" class="loading-spinner"></div>
        <div v-else>
          <div v-if="output" class="output">
            <h4>Output:</h4>
            <div v-html="formatOutput(output)"></div>
          </div>
          <div v-if="details" class="output">
            <h4>Details:</h4>
            <div v-html="formatOutput(details)"></div>
          </div>
          <div v-if="dockerDetails" class="output">
            <h4>Docker Details:</h4>
            <div v-html="formatOutput(dockerDetails)"></div>
          </div>
        </div>
      </div>
    </main>
    <footer>
      <p>Please consider donating to allow sites like these to keep running!</p>    
      <form action="https://www.paypal.com/donate" method="post" target="_top">
        <input type="hidden" name="business" value="R4KZAGAML2VMS" />
        <input type="hidden" name="no_recurring" value="0" />
        <input type="hidden" name="item_name" value="Thank you for donating! This will help me create new projects, and keep current ones running!" />
        <input type="hidden" name="currency_code" value="USD" />
        <input type="image" src="https://www.paypalobjects.com/en_US/i/btn/btn_donate_SM.gif" border="0" name="submit" title="PayPal - The safer, easier way to pay online!" alt="Donate with PayPal button" />
        <img alt="" border="0" src="https://www.paypal.com/en_US/i/scr/pixel.gif" width="1" height="1" />
      </form>
      <p>Any issues with this app may be reported <a href="https://github.com/DerekCorniello/pip-req-valid/issues">here</a>.</p>
    </footer>
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
      dockerDetails: null,
      authToken: null, // Store the token
    };
  },
  methods: {
    async getAuthToken() {
      try {
        const response = await fetch("https://api.reqinspect.com/auth", {
          method: "GET",
        });

        if (response.ok) {
          const jsonResponse = await response.json();
          this.authToken = jsonResponse.token;
        } else if (response.status == 429) {
          this.output = "Server Rate Limit Exceeded, Please check back soon!"	
        } else {
          this.output = "Error obtaining authentication token.";
        }
      } catch (error) {
        this.output = `An error occurred while obtaining the token: ${error.message}`;
      }
    },

    async handleFileSubmission(file) {
      this.loading = true;
      this.output = null;
      this.dockerDetails = null;
      await this.getAuthToken();

      if (!this.authToken) {
        this.loading = false;
        return;
      }

      try {
        const formData = new FormData();
        formData.append("file", file);

        const response = await fetch("https://api.reqinspect.com", {
          method: "POST",
          headers: {
            "Authorization": `Bearer ${this.authToken}`,
          },
          body: formData,
        });

        if (response.ok) {
          const jsonResponse = await response.json();
          this.output = jsonResponse.prettyOutput;
          this.dockerDetails = jsonResponse.installOutput;

          this.details = 
            jsonResponse.details && jsonResponse.errors
              ? jsonResponse.details + "\n" + jsonResponse.errors
              : jsonResponse.details || jsonResponse.errors || "No details found";
        } else if (response.status == 429) {
          this.output = "Server Rate Limit Exceeded, Please check back soon!"	
        } else {
          this.output = "Error obtaining authentication token.";
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

    html, body {
      overscroll-behavior: none;
    }

    body {
      margin: 0;
      font-family: Arial, sans-serif;
      background-color: #121212;
      color: #e0e0e0;
      width: 100%;
    }

    #app {
      display: flex;
      flex-direction: column;
      align-items: center;
      text-align: center;
      min-height: 100vh;
      width: 100%;
    }

    header h1 {
      margin: 1rem 0;
      font-size: 2rem;
    }

    main {
      flex: 1;
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

    footer {
      background-color: #1e1e1e;
      text-align: center;
      border-top: 1px solid #ddd;
      position: sticky;
      bottom: 0;
      width: 100%;
    }

    footer a {
      color: #007bff;
      text-decoration: none;
    }

    footer a:hover {
      text-decoration: underline;
    }

</style>
