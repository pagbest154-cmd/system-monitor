package ru.ferrumnst.sysmon.data.api

import ru.ferrumnst.sysmon.data.models.MetricsHistoryResponse
import ru.ferrumnst.sysmon.data.session.HubSession
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.Request
import java.net.URLEncoder

object MetricsFetcher {
    suspend fun fetchHistory(
        session: HubSession,
        sensorId: String,
        agentId: String?,
        period: String,
    ): MetricsHistoryResponse = withContext(Dispatchers.IO) {
        val base = session.hubUrl.trim().removeSuffix("/")
        val encodedSensor = URLEncoder.encode(sensorId, Charsets.UTF_8.name())
        val encodedPeriod = URLEncoder.encode(period, Charsets.UTF_8.name())
        val agentPart = if (!agentId.isNullOrBlank()) {
            "&agent=${URLEncoder.encode(agentId, Charsets.UTF_8.name())}"
        } else {
            ""
        }
        val url = "$base/api/metrics/$encodedSensor?period=$encodedPeriod$agentPart"

        val requestBuilder = Request.Builder()
            .url(url)
            .get()

        if (session.hasCredentials) {
            val credentials = okhttp3.Credentials.basic(session.hubName, session.hubKey)
            requestBuilder.header("Authorization", credentials)
        }

        val client = HubClientFactory.createOkHttpClient(session)
        val response = client.newCall(requestBuilder.build()).execute()
        val body = response.body?.string() ?: throw IllegalStateException("Пустой ответ metrics API")
        if (!response.isSuccessful) {
            throw IllegalStateException("metrics API ${response.code}: $body")
        }
        HubClientFactory.json.decodeFromString(MetricsHistoryResponse.serializer(), body)
    }
}
