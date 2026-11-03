plugins { id("com.android.application") }
android {
    namespace = "kr.pulsenote.pn12usb"
    compileSdk = 35
    defaultConfig {
        applicationId = "kr.pulsenote.pn12usb"
        minSdk = 26
        targetSdk = 35
        versionCode = 21
        versionName = "1.3.5"
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}
