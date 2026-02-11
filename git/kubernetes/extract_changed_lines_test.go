package kubernetes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gkubernetes "github.com/flanksource/gavel/git/kubernetes"
)

var _ = Describe("ExtractChangedLines", func() {
	Describe("Pure Addition", func() {
		It("should track only added lines", func() {
			patch := `diff --git a/file.yaml b/file.yaml
index abc123..def456 100644
--- a/file.yaml
+++ b/file.yaml
@@ -1,3 +1,4 @@
 line1
 line2
+line3
 line4`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(Equal([]int{3}))
			Expect(result.DeletedLines).To(BeEmpty())
		})
	})

	Describe("Pure Deletion", func() {
		It("should track only deleted lines", func() {
			patch := `diff --git a/file.yaml b/file.yaml
index abc123..def456 100644
--- a/file.yaml
+++ b/file.yaml
@@ -1,4 +1,3 @@
 line1
-line2
 line3
 line4`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(BeEmpty())
			Expect(result.DeletedLines).To(Equal([]int{2}))
		})
	})

	Describe("Modification", func() {
		It("should track both deletion and addition at same position", func() {
			patch := `diff --git a/deployment.yaml b/deployment.yaml
index abc123..def456 100644
--- a/deployment.yaml
+++ b/deployment.yaml
@@ -4,7 +4,7 @@ metadata:
   name: nginx-deployment
   namespace: default
 spec:
-  replicas: 2
+  replicas: 5
   selector:
     matchLabels:
       app: nginx`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(Equal([]int{7}))
			Expect(result.DeletedLines).To(Equal([]int{7}))
		})
	})

	Describe("Line Movement", func() {
		It("should track deletion and addition at different positions", func() {
			patch := `diff --git a/file.yaml b/file.yaml
index abc123..def456 100644
--- a/file.yaml
+++ b/file.yaml
@@ -8,4 +8,3 @@
   data:
     config.yaml: |
       setting1: value1
-      setting2: value2
       setting3: value3
@@ -22,3 +21,4 @@
     database:
       host: localhost
       port: 5432
+      setting2: value2`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.DeletedLines).To(ContainElement(11))
			Expect(result.AddedLines).To(ContainElement(24))
		})
	})

	Describe("Multiple Additions and Deletions", func() {
		It("should track all changed lines", func() {
			patch := `diff --git a/file.yaml b/file.yaml
index abc123..def456 100644
--- a/file.yaml
+++ b/file.yaml
@@ -1,6 +1,7 @@
 line1
-line2
+line2-modified
 line3
+line4-new
 line5
-line6
 line7`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(ConsistOf(2, 4))
			Expect(result.DeletedLines).To(ConsistOf(2, 5))
		})
	})

	Describe("Multiple Hunks", func() {
		It("should handle multiple @@ sections correctly", func() {
			patch := `diff --git a/file.yaml b/file.yaml
index abc123..def456 100644
--- a/file.yaml
+++ b/file.yaml
@@ -5,7 +5,7 @@ metadata:
   name: app
 spec:
-  replicas: 2
+  replicas: 3
   selector:
@@ -20,6 +20,7 @@ spec:
     spec:
       containers:
       - name: nginx
+        image: nginx:1.19
         ports:
         - containerPort: 80`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(ConsistOf(7, 23))
			Expect(result.DeletedLines).To(Equal([]int{7}))
		})
	})

	Describe("Empty Patch", func() {
		It("should return empty slices", func() {
			patch := ``

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(BeEmpty())
			Expect(result.DeletedLines).To(BeEmpty())
		})
	})

	Describe("Context-Only Patch", func() {
		It("should return empty slices when only context lines present", func() {
			patch := `diff --git a/file.yaml b/file.yaml
index abc123..def456 100644
--- a/file.yaml
+++ b/file.yaml
@@ -1,3 +1,3 @@
 line1
 line2
 line3`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(BeEmpty())
			Expect(result.DeletedLines).To(BeEmpty())
		})
	})

	Describe("Full File Deletion", func() {
		It("should track all deleted lines", func() {
			patch := `diff --git a/file.yaml b/file.yaml
deleted file mode 100644
index abc123..0000000
--- a/file.yaml
+++ /dev/null
@@ -1,5 +0,0 @@
-apiVersion: v1
-kind: ConfigMap
-metadata:
-  name: my-config
-  namespace: default`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(BeEmpty())
			Expect(result.DeletedLines).To(ConsistOf(1, 2, 3, 4, 5))
		})
	})

	Describe("Full File Addition", func() {
		It("should track all added lines", func() {
			patch := `diff --git a/file.yaml b/file.yaml
new file mode 100644
index 0000000..abc123
--- /dev/null
+++ b/file.yaml
@@ -0,0 +1,5 @@
+apiVersion: v1
+kind: ConfigMap
+metadata:
+  name: my-config
+  namespace: default`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(ConsistOf(1, 2, 3, 4, 5))
			Expect(result.DeletedLines).To(BeEmpty())
		})
	})

	Describe("Complex Multi-Document Change", func() {
		It("should handle changes across multiple YAML documents", func() {
			patch := `diff --git a/multi.yaml b/multi.yaml
index abc123..def456 100644
--- a/multi.yaml
+++ b/multi.yaml
@@ -8,7 +8,7 @@ metadata:
 data:
   key1: value1
-  key2: value2
+  key2: value2-updated
 ---
 apiVersion: v1
@@ -20,6 +20,8 @@ metadata:
 data:
   setting1: config1
+  setting2: config2
+  setting3: config3
   setting4: config4`

			result := gkubernetes.ExtractChangedLines(patch)
			Expect(result.AddedLines).To(ConsistOf(10, 22, 23))
			Expect(result.DeletedLines).To(Equal([]int{10}))
		})
	})
})
