# PulseNote Deck website archive

This folder is for the PulseNote Deck PN15 website only. It does **not** modify the Windows application, firmware or any release files.

Upload the complete website archive as `site.zip` in this directory. Do **not** extract the ZIP in GitHub. The separate GitHub Actions workflow validates and publishes the original HTML/CSS/JavaScript and all **15 unmodified images** from the archive.

## Publishing
1. Visit [this folder](https://github.com/neoartmin/pulsenote-deck/tree/main/website) and choose **Add file → Upload files**.
2. Upload `site.zip` and commit on `main`.
3. Under [Settings → Pages](https://github.com/neoartmin/pulsenote-deck/settings/pages), select **GitHub Actions** as the build and deployment source.
4. See [Publish PulseNote Deck website](https://github.com/neoartmin/pulsenote-deck/actions/workflows/deploy-pulsenote-site.yml) for deployment results. If needed, choose **Run workflow**.
5. Upon success, visit https://neoartmin.github.io/pulsenote-deck/.

Replacing `website/site.zip` triggers an automatic site update. All images retain their original bytes. 

Current site package: **PulseNote_Deck_Web_v0.2.zip**.
