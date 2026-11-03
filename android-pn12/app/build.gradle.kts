plugins { id("com.android.application") }
android {
    namespace = "kr.pulsenote.pn12usb"
    compileSdk = 35
    defaultConfig {
        applicationId = "kr.pulsenote.pn12usb"
        minSdk = 26
        targetSdk = 35
        versionCode = 25
        versionName = "1.3.9"
    }
    buildFeatures {
        buildConfig = true
    }
    flavorDimensions += "security"
    productFlavors {
        create("standard") {
            dimension = "security"
            // USB/media/screenshot only. Remove sensitive manifest components.
            buildConfigField("boolean", "HAS_ACCESSIBILITY", "false")
            buildConfigField("boolean", "HAS_OVERLAY", "false")
        }
        create("extended") {
            dimension = "security"
            // Opt-in advanced build retains existing accessibility/overlay features.
            buildConfigField("boolean", "HAS_ACCESSIBILITY", "true")
            buildConfigField("boolean", "HAS_OVERLAY", "true")
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}
