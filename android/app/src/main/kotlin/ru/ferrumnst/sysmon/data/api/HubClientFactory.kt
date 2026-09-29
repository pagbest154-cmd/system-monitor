package ru.ferrumnst.sysmon.data.api

import ru.ferrumnst.sysmon.data.session.HubSession
import kotlinx.serialization.json.Json
import okhttp3.Credentials
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.kotlinx.serialization.asConverterFactory
import okhttp3.MediaType.Companion.toMediaType
import java.util.concurrent.TimeUnit

object HubClientFactory {
    val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        coerceInputValues = true
    }

    fun createApi(session: HubSession, debugLogging: Boolean = false): HubApi {
        val baseUrl = normalizeBaseUrl(session.hubUrl)
        val client = createOkHttpClient(session, debugLogging)
        return Retrofit.Builder()
            .baseUrl(baseUrl)
            .client(client)
            .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
            .build()
            .create(HubApi::class.java)
    }

    fun createOkHttpClient(session: HubSession, debugLogging: Boolean = false): OkHttpClient {
        val builder = OkHttpClient.Builder()
            .connectTimeout(15, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)

        if (session.hasCredentials) {
            val authHeader = Credentials.basic(session.hubName, session.hubKey)
            builder.addInterceptor { chain ->
                chain.proceed(
                    chain.request().newBuilder()
                        .header("Authorization", authHeader)
                        .build(),
                )
            }
        }

        if (debugLogging) {
            builder.addInterceptor(
                HttpLoggingInterceptor().apply {
                    level = HttpLoggingInterceptor.Level.BASIC
                },
            )
        }

        return builder.build()
    }

    fun normalizeBaseUrl(url: String): String {
        val trimmed = url.trim().removeSuffix("/")
        return if (trimmed.endsWith("/")) trimmed else "$trimmed/"
    }

    fun webSocketUrl(session: HubSession, agentId: String?): String {
        val base = session.hubUrl.trim().removeSuffix("/")
        val wsBase = when {
            base.startsWith("https://") -> "wss://" + base.removePrefix("https://")
            base.startsWith("http://") -> "ws://" + base.removePrefix("http://")
            else -> "http://$base".replace("http://", "ws://")
        }
        val path = "/ws/live"
        return if (agentId.isNullOrBlank()) {
            "$wsBase$path"
        } else {
            "$wsBase$path?agent=${java.net.URLEncoder.encode(agentId, Charsets.UTF_8.name())}"
        }
    }
}
