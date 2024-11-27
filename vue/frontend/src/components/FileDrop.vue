<template>
  <div class="file-drop" @dragover.prevent @drop.prevent="handleDrop">
    <p>Drag and drop your requirements.txt file here or click below to upload.</p>
    <p class="subtext">(.txt only, 5 MB limit)</p>
    <input type="file" ref="fileInput" @change="handleFileInput" hidden />
    <button @click="triggerFileInput">Browse File</button><br><br>
    <div v-if="fileName" class="file-info">Selected file: {{ fileName }}</div>
    <div v-if="error" class="error">{{ error }}</div><br>
    <button v-if="file" @click="submitFile">Validate</button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      file: null,
      fileName: "",
      error: "",
    };
  },
  methods: {
    handleDrop(event) {
      const droppedFile = event.dataTransfer.files[0];
      this.validateFile(droppedFile);
    },
    handleFileInput(event) {
      const selectedFile = event.target.files[0];
      this.validateFile(selectedFile);
    },
    triggerFileInput() {
      this.$refs.fileInput.click();
    },
    validateFile(file) {
      if (!file) return;

      const validExtension = file.name.endsWith(".txt");
      const validSize = file.size <= 5 * 1024 * 1024; // 5 MB

      if (!validExtension) {
        this.error = "Only .txt files are allowed.";
        this.resetFile();
        return;
      }
      if (!validSize) {
        this.error = "File size must not exceed 5 MB.";
        this.resetFile();
        return;
      }

      this.file = file;
      this.fileName = file.name;
      this.error = "";
    },
    submitFile() {
      if (this.file) {
        this.$emit("file-submitted", this.file);
      }
    },
    resetFile() {
      this.file = null;
      this.fileName = "";
    },
  },
};
</script>

<style>
.file-drop {
  border: 2px dashed #444;
  padding: 2rem;
  border-radius: 8px;
  text-align: center;
  background-color: #1e1e1e;
  cursor: pointer;
  transition: border-color 0.3s;
}

.file-drop:hover {
  border-color: #6200ea;
}

.file-drop button {
  margin-top: 1rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  background-color: #6200ea;
  color: white;
  cursor: pointer;
}

.file-drop button:hover {
  background-color: #3700b3;
}

.file-drop .subtext {
  font-size: 0.9rem;
  color: #bbb;
}

.file-info {
  margin-top: 1rem;
  font-size: 0.95rem;
  color: #e0e0e0;
}

.error {
  margin-top: 1rem;
  font-size: 0.9rem;
  color: #f44336;
}
</style>
