--
-- PostgreSQL database dump
--

\restrict sLw193cihpID2OBD9TtbiMEBIdmc8hwgcSXeeE144gmcSkkufWdKYRJRKoEwRju

-- Dumped from database version 15.17
-- Dumped by pg_dump version 15.17

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: fn_update_artist_song_counters_deletion(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_update_artist_song_counters_deletion() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (OLD.status = TRUE) THEN
                    UPDATE artists SET enabled_songs = GREATEST(0, enabled_songs - 1)
                    WHERE id IN (SELECT artist_id FROM artist_song WHERE song_id = OLD.id);
                ELSE
                    UPDATE artists SET disabled_songs = GREATEST(0, disabled_songs - 1)
                    WHERE id IN (SELECT artist_id FROM artist_song WHERE song_id = OLD.id);
                END IF;
                RETURN OLD;
            END;
            $$;


ALTER FUNCTION public.fn_update_artist_song_counters_deletion() OWNER TO postgres;

--
-- Name: fn_update_artist_song_counters_pivot(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_update_artist_song_counters_pivot() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            DECLARE
                song_status BOOLEAN;
            BEGIN
                IF (TG_OP = 'INSERT') THEN
                    SELECT status INTO song_status FROM songs WHERE id = NEW.song_id;
                    IF (song_status = TRUE) THEN
                        UPDATE artists SET enabled_songs = enabled_songs + 1 WHERE id = NEW.artist_id;
                    ELSE
                        UPDATE artists SET disabled_songs = disabled_songs + 1 WHERE id = NEW.artist_id;
                    END IF;
                ELSIF (TG_OP = 'DELETE') THEN
                    SELECT status INTO song_status FROM songs WHERE id = OLD.song_id;
                    IF (song_status IS NOT NULL) THEN
                        IF (song_status = TRUE) THEN
                            UPDATE artists SET enabled_songs = GREATEST(0, enabled_songs - 1) WHERE id = OLD.artist_id;
                        ELSE
                            UPDATE artists SET disabled_songs = GREATEST(0, disabled_songs - 1) WHERE id = OLD.artist_id;
                        END IF;
                    END IF;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.fn_update_artist_song_counters_pivot() OWNER TO postgres;

--
-- Name: fn_update_artist_song_counters_status(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_update_artist_song_counters_status() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (OLD.status = FALSE AND NEW.status = TRUE) THEN
                    UPDATE artists SET
                        enabled_songs = enabled_songs + 1,
                        disabled_songs = GREATEST(0, disabled_songs - 1)
                    WHERE id IN (SELECT artist_id FROM artist_song WHERE song_id = NEW.id);
                ELSIF (OLD.status = TRUE AND NEW.status = FALSE) THEN
                    UPDATE artists SET
                        enabled_songs = GREATEST(0, enabled_songs - 1),
                        disabled_songs = disabled_songs + 1
                    WHERE id IN (SELECT artist_id FROM artist_song WHERE song_id = NEW.id);
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.fn_update_artist_song_counters_status() OWNER TO postgres;

--
-- Name: fn_update_daily_song_views(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_update_daily_song_views() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                INSERT INTO daily_metrics (song_id, date, views_count, created_at, updated_at)
                VALUES (NEW.id, CURRENT_DATE, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ON CONFLICT (song_id, date) 
                DO UPDATE SET 
                    views_count = daily_metrics.views_count + 1,
                    updated_at = CURRENT_TIMESTAMP;
                RETURN NEW;
            END;
            $$;


ALTER FUNCTION public.fn_update_daily_song_views() OWNER TO postgres;

--
-- Name: fn_update_daily_variant_views(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_update_daily_variant_views() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                INSERT INTO daily_metrics (song_id, date, views_count, created_at, updated_at)
                VALUES (NEW.song_id, CURRENT_DATE, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                ON CONFLICT (song_id, date) 
                DO UPDATE SET 
                    views_count = daily_metrics.views_count + 1,
                    updated_at = CURRENT_TIMESTAMP;
                RETURN NEW;
            END;
            $$;


ALTER FUNCTION public.fn_update_daily_variant_views() OWNER TO postgres;

--
-- Name: fn_update_producer_anime_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_update_producer_anime_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (TG_OP = 'INSERT') THEN
                    UPDATE producers SET anime_count = anime_count + 1 WHERE id = NEW.producer_id;
                ELSIF (TG_OP = 'DELETE') THEN
                    UPDATE producers SET anime_count = GREATEST(0, anime_count - 1) WHERE id = OLD.producer_id;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.fn_update_producer_anime_count() OWNER TO postgres;

--
-- Name: fn_update_studio_anime_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_update_studio_anime_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (TG_OP = 'INSERT') THEN
                    UPDATE studios SET anime_count = anime_count + 1 WHERE id = NEW.studio_id;
                ELSIF (TG_OP = 'DELETE') THEN
                    UPDATE studios SET anime_count = GREATEST(0, anime_count - 1) WHERE id = OLD.studio_id;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.fn_update_studio_anime_count() OWNER TO postgres;

--
-- Name: handle_artist_song_change(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.handle_artist_song_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
                    PERFORM recount_artist_stats(NEW.artist_id);
                END IF;
                IF (TG_OP = 'DELETE' OR TG_OP = 'UPDATE') THEN
                    PERFORM recount_artist_stats(OLD.artist_id);
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.handle_artist_song_change() OWNER TO postgres;

--
-- Name: handle_song_status_change(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.handle_song_status_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            DECLARE
                r RECORD;
            BEGIN
                -- Only trigger if status changed
                IF (OLD.status IS DISTINCT FROM NEW.status) THEN
                    FOR r IN SELECT artist_id FROM artist_song WHERE song_id = NEW.id LOOP
                        PERFORM recount_artist_stats(r.artist_id);
                    END LOOP;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.handle_song_status_change() OWNER TO postgres;

--
-- Name: recount_artist_stats(bigint); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.recount_artist_stats(artist_id_param bigint) RETURNS void
    LANGUAGE plpgsql
    AS $$
            BEGIN
                UPDATE artists SET
                    enabled_songs = (
                        SELECT COUNT(*) 
                        FROM artist_song 
                        JOIN songs ON artist_song.song_id = songs.id 
                        WHERE artist_song.artist_id = artist_id_param AND songs.status = true
                    ),
                    disabled_songs = (
                        SELECT COUNT(*) 
                        FROM artist_song 
                        JOIN songs ON artist_song.song_id = songs.id 
                        WHERE artist_song.artist_id = artist_id_param AND songs.status = false
                    )
                WHERE id = artist_id_param;
            END;
            $$;


ALTER FUNCTION public.recount_artist_stats(artist_id_param bigint) OWNER TO postgres;

--
-- Name: update_anime_song_counts(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_anime_song_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (TG_OP = 'INSERT') THEN
                    IF NEW.status THEN
                        UPDATE animes SET enabled_songs = enabled_songs + 1 WHERE id = NEW.anime_id;
                    ELSE
                        UPDATE animes SET disabled_songs = disabled_songs + 1 WHERE id = NEW.anime_id;
                    END IF;
                ELSIF (TG_OP = 'UPDATE') THEN
                    -- Case 1: Same anime, status changed
                    IF (OLD.anime_id = NEW.anime_id) AND (OLD.status != NEW.status) THEN
                        IF NEW.status THEN
                            UPDATE animes SET enabled_songs = enabled_songs + 1, disabled_songs = disabled_songs - 1 WHERE id = NEW.anime_id;
                        ELSE
                            UPDATE animes SET enabled_songs = enabled_songs - 1, disabled_songs = disabled_songs + 1 WHERE id = NEW.anime_id;
                        END IF;
                    -- Case 2: Anime changed
                    ELSIF (OLD.anime_id != NEW.anime_id) THEN
                        -- Decrement old anime
                        IF OLD.status THEN
                            UPDATE animes SET enabled_songs = enabled_songs - 1 WHERE id = OLD.anime_id;
                        ELSE
                            UPDATE animes SET disabled_songs = disabled_songs - 1 WHERE id = OLD.anime_id;
                        END IF;
                        -- Increment new anime
                        IF NEW.status THEN
                            UPDATE animes SET enabled_songs = enabled_songs + 1 WHERE id = NEW.anime_id;
                        ELSE
                            UPDATE animes SET disabled_songs = disabled_songs + 1 WHERE id = NEW.anime_id;
                        END IF;
                    END IF;
                ELSIF (TG_OP = 'DELETE') THEN
                    IF OLD.status THEN
                        UPDATE animes SET enabled_songs = enabled_songs - 1 WHERE id = OLD.anime_id;
                    ELSE
                        UPDATE animes SET disabled_songs = disabled_songs - 1 WHERE id = OLD.anime_id;
                    END IF;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.update_anime_song_counts() OWNER TO postgres;

--
-- Name: update_anime_songs_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_anime_songs_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (TG_OP = 'INSERT') THEN
                    UPDATE animes SET songs_count = songs_count + 1 WHERE id = NEW.anime_id;
                ELSIF (TG_OP = 'DELETE') THEN
                    UPDATE animes SET songs_count = GREATEST(0, songs_count - 1) WHERE id = OLD.anime_id;
                ELSIF (TG_OP = 'UPDATE' AND (OLD.anime_id IS DISTINCT FROM NEW.anime_id)) THEN
                    IF (OLD.anime_id IS NOT NULL) THEN
                        UPDATE animes SET songs_count = GREATEST(0, songs_count - 1) WHERE id = OLD.anime_id;
                    END IF;
                    IF (NEW.anime_id IS NOT NULL) THEN
                        UPDATE animes SET songs_count = songs_count + 1 WHERE id = NEW.anime_id;
                    END IF;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.update_anime_songs_count() OWNER TO postgres;

--
-- Name: update_artist_favorites_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_artist_favorites_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (TG_OP = 'INSERT') THEN
                    UPDATE artists SET favorites_count = favorites_count + 1 WHERE id = NEW.artist_id;
                ELSIF (TG_OP = 'DELETE') THEN
                    UPDATE artists SET favorites_count = favorites_count - 1 WHERE id = OLD.artist_id;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.update_artist_favorites_count() OWNER TO postgres;

--
-- Name: update_song_average_score(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_song_average_score() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
                    UPDATE songs SET average_score = (SELECT AVG(rating) FROM song_ratings WHERE song_id = NEW.song_id) WHERE id = NEW.song_id;
                ELSIF (TG_OP = 'DELETE') THEN
                    UPDATE songs SET average_score = COALESCE((SELECT AVG(rating) FROM song_ratings WHERE song_id = OLD.song_id), 0) WHERE id = OLD.song_id;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.update_song_average_score() OWNER TO postgres;

--
-- Name: update_song_favorites_count(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_song_favorites_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            BEGIN
                IF (TG_OP = 'INSERT') THEN
                    UPDATE songs SET favorites_count = favorites_count + 1 WHERE id = NEW.song_id;
                ELSIF (TG_OP = 'DELETE') THEN
                    UPDATE songs SET favorites_count = favorites_count - 1 WHERE id = OLD.song_id;
                END IF;
                RETURN NULL;
            END;
            $$;


ALTER FUNCTION public.update_song_favorites_count() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: activities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.activities (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    action_type character varying(50) NOT NULL,
    target_id bigint NOT NULL,
    target_type character varying(191) NOT NULL,
    action_value text,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.activities OWNER TO postgres;

--
-- Name: activities_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.activities_id_seq OWNER TO postgres;

--
-- Name: activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.activities_id_seq OWNED BY public.activities.id;


--
-- Name: anime_external_link; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.anime_external_link (
    id bigint NOT NULL,
    anime_id bigint NOT NULL,
    external_link_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.anime_external_link OWNER TO postgres;

--
-- Name: anime_genre; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.anime_genre (
    id bigint NOT NULL,
    genre_id bigint NOT NULL,
    anime_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.anime_genre OWNER TO postgres;

--
-- Name: anime_producer; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.anime_producer (
    id bigint NOT NULL,
    anime_id bigint NOT NULL,
    producer_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.anime_producer OWNER TO postgres;

--
-- Name: anime_studio; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.anime_studio (
    id bigint NOT NULL,
    anime_id bigint NOT NULL,
    studio_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.anime_studio OWNER TO postgres;

--
-- Name: animes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.animes (
    id bigint NOT NULL,
    title character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    description text,
    anilist_id bigint,
    status boolean DEFAULT false NOT NULL,
    year_id bigint,
    season_id bigint,
    format_id bigint,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    cover character varying(191),
    banner character varying(191),
    songs_count integer DEFAULT 0 NOT NULL,
    uuid uuid,
    anime_themes_id bigint,
    enabled_songs integer DEFAULT 0 NOT NULL,
    disabled_songs integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.animes OWNER TO postgres;

--
-- Name: announcements; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.announcements (
    id bigint NOT NULL,
    title character varying(191) NOT NULL,
    content text,
    type character varying(191) DEFAULT 'info'::character varying NOT NULL,
    icon character varying(191),
    url character varying(191),
    image character varying(191),
    priority integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    starts_at timestamp(0) without time zone,
    ends_at timestamp(0) without time zone,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.announcements OWNER TO postgres;

--
-- Name: announcements_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.announcements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.announcements_id_seq OWNER TO postgres;

--
-- Name: announcements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.announcements_id_seq OWNED BY public.announcements.id;


--
-- Name: artist_song; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artist_song (
    id bigint NOT NULL,
    artist_id bigint NOT NULL,
    song_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.artist_song OWNER TO postgres;

--
-- Name: artist_song_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.artist_song_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.artist_song_id_seq OWNER TO postgres;

--
-- Name: artist_song_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.artist_song_id_seq OWNED BY public.artist_song.id;


--
-- Name: artist_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artist_user (
    id bigint NOT NULL,
    artist_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.artist_user OWNER TO postgres;

--
-- Name: artist_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.artist_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.artist_user_id_seq OWNER TO postgres;

--
-- Name: artist_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.artist_user_id_seq OWNED BY public.artist_user.id;


--
-- Name: artists; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artists (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    name_jp character varying(191),
    slug character varying(191) NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    avatar character varying(191),
    status boolean DEFAULT false NOT NULL,
    favorites_count integer DEFAULT 0 NOT NULL,
    uuid uuid,
    anilist_id bigint,
    anime_themes_id bigint,
    enabled_songs integer DEFAULT 0 NOT NULL,
    disabled_songs integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.artists OWNER TO postgres;

--
-- Name: artists_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.artists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.artists_id_seq OWNER TO postgres;

--
-- Name: artists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.artists_id_seq OWNED BY public.artists.id;


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.audit_logs (
    id bigint NOT NULL,
    user_id bigint,
    event character varying(50) NOT NULL,
    auditable_id bigint NOT NULL,
    auditable_type character varying(120) NOT NULL,
    old_values json,
    new_values json,
    url text,
    ip_address character varying(45),
    user_agent text,
    created_at timestamp(0) without time zone
);


ALTER TABLE public.audit_logs OWNER TO postgres;

--
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.audit_logs_id_seq OWNER TO postgres;

--
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;


--
-- Name: badge_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.badge_user (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    badge_id bigint NOT NULL,
    awarded_at timestamp(0) without time zone,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.badge_user OWNER TO postgres;

--
-- Name: badge_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.badge_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.badge_user_id_seq OWNER TO postgres;

--
-- Name: badge_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.badge_user_id_seq OWNED BY public.badge_user.id;


--
-- Name: badges; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.badges (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    description text,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    icon character varying(191)
);


ALTER TABLE public.badges OWNER TO postgres;

--
-- Name: badges_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.badges_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.badges_id_seq OWNER TO postgres;

--
-- Name: badges_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.badges_id_seq OWNED BY public.badges.id;


--
-- Name: cache; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cache (
    key character varying(191) NOT NULL,
    value text NOT NULL,
    expiration integer NOT NULL
);


ALTER TABLE public.cache OWNER TO postgres;

--
-- Name: cache_locks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cache_locks (
    key character varying(191) NOT NULL,
    owner character varying(191) NOT NULL,
    expiration integer NOT NULL
);


ALTER TABLE public.cache_locks OWNER TO postgres;

--
-- Name: comment_reactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.comment_reactions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    comment_id bigint NOT NULL,
    type smallint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.comment_reactions OWNER TO postgres;

--
-- Name: comment_reactions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.comment_reactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.comment_reactions_id_seq OWNER TO postgres;

--
-- Name: comment_reactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.comment_reactions_id_seq OWNED BY public.comment_reactions.id;


--
-- Name: comment_reports; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.comment_reports (
    id bigint NOT NULL,
    comment_id bigint NOT NULL,
    user_id bigint NOT NULL,
    title character varying(191) NOT NULL,
    content text,
    source character varying(191),
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    status boolean DEFAULT false NOT NULL
);


ALTER TABLE public.comment_reports OWNER TO postgres;

--
-- Name: comment_reports_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.comment_reports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.comment_reports_id_seq OWNER TO postgres;

--
-- Name: comment_reports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.comment_reports_id_seq OWNED BY public.comment_reports.id;


--
-- Name: comments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.comments (
    id bigint NOT NULL,
    parent_id bigint,
    user_id bigint NOT NULL,
    content text NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    song_id bigint,
    likes_count bigint DEFAULT '0'::bigint NOT NULL,
    dislikes_count bigint DEFAULT '0'::bigint NOT NULL
);


ALTER TABLE public.comments OWNER TO postgres;

--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.comments_id_seq OWNER TO postgres;

--
-- Name: comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.comments_id_seq OWNED BY public.comments.id;


--
-- Name: daily_metrics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.daily_metrics (
    id bigint NOT NULL,
    song_id bigint,
    date date NOT NULL,
    views_count integer DEFAULT 0 NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    new_users_count integer DEFAULT 0 NOT NULL,
    new_ratings_count integer DEFAULT 0 NOT NULL,
    new_songs_count integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.daily_metrics OWNER TO postgres;

--
-- Name: daily_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.daily_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.daily_metrics_id_seq OWNER TO postgres;

--
-- Name: daily_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.daily_metrics_id_seq OWNED BY public.daily_metrics.id;


--
-- Name: external_link_post_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.external_link_post_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.external_link_post_id_seq OWNER TO postgres;

--
-- Name: external_link_post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.external_link_post_id_seq OWNED BY public.anime_external_link.id;


--
-- Name: external_links; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.external_links (
    id bigint NOT NULL,
    icon character varying(191),
    type character varying(191) NOT NULL,
    name character varying(191) NOT NULL,
    url character varying(191) NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.external_links OWNER TO postgres;

--
-- Name: external_links_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.external_links_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.external_links_id_seq OWNER TO postgres;

--
-- Name: external_links_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.external_links_id_seq OWNED BY public.external_links.id;


--
-- Name: failed_jobs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.failed_jobs (
    id bigint NOT NULL,
    uuid character varying(191) NOT NULL,
    connection text NOT NULL,
    queue text NOT NULL,
    payload text NOT NULL,
    exception text NOT NULL,
    failed_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.failed_jobs OWNER TO postgres;

--
-- Name: failed_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.failed_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.failed_jobs_id_seq OWNER TO postgres;

--
-- Name: failed_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.failed_jobs_id_seq OWNED BY public.failed_jobs.id;


--
-- Name: follows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.follows (
    follower_id bigint NOT NULL,
    followed_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.follows OWNER TO postgres;

--
-- Name: formats; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.formats (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.formats OWNER TO postgres;

--
-- Name: formats_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.formats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.formats_id_seq OWNER TO postgres;

--
-- Name: formats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.formats_id_seq OWNED BY public.formats.id;


--
-- Name: genre_post_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.genre_post_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.genre_post_id_seq OWNER TO postgres;

--
-- Name: genre_post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.genre_post_id_seq OWNED BY public.anime_genre.id;


--
-- Name: genres; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.genres (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.genres OWNER TO postgres;

--
-- Name: genres_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.genres_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.genres_id_seq OWNER TO postgres;

--
-- Name: genres_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.genres_id_seq OWNED BY public.genres.id;


--
-- Name: levels; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.levels (
    level integer NOT NULL,
    min_xp bigint NOT NULL,
    name character varying(50),
    badge_id bigint,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.levels OWNER TO postgres;

--
-- Name: migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.migrations (
    id integer NOT NULL,
    migration character varying(191) NOT NULL,
    batch integer NOT NULL
);


ALTER TABLE public.migrations OWNER TO postgres;

--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.migrations_id_seq OWNER TO postgres;

--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.migrations_id_seq OWNED BY public.migrations.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notifications (
    id uuid NOT NULL,
    user_id bigint NOT NULL,
    type character varying(50) NOT NULL,
    subject_id bigint,
    subject_type character varying(50),
    data json NOT NULL,
    read_at timestamp(0) without time zone,
    deleted_at timestamp(0) without time zone,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    subject_uuid uuid
);


ALTER TABLE public.notifications OWNER TO postgres;

--
-- Name: password_resets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.password_resets (
    email character varying(191) NOT NULL,
    token character varying(191) NOT NULL,
    created_at timestamp(0) without time zone
);


ALTER TABLE public.password_resets OWNER TO postgres;

--
-- Name: permissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.permissions (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    description character varying(191),
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.permissions OWNER TO postgres;

--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.permissions_id_seq OWNER TO postgres;

--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: personal_access_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.personal_access_tokens (
    id bigint NOT NULL,
    tokenable_type character varying(191) NOT NULL,
    tokenable_id bigint NOT NULL,
    name character varying(191) NOT NULL,
    token character varying(64) NOT NULL,
    abilities text,
    last_used_at timestamp(0) without time zone,
    expires_at timestamp(0) without time zone,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.personal_access_tokens OWNER TO postgres;

--
-- Name: personal_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.personal_access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.personal_access_tokens_id_seq OWNER TO postgres;

--
-- Name: personal_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.personal_access_tokens_id_seq OWNED BY public.personal_access_tokens.id;


--
-- Name: playlist_song; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.playlist_song (
    id bigint NOT NULL,
    playlist_id bigint NOT NULL,
    song_id bigint NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.playlist_song OWNER TO postgres;

--
-- Name: playlist_song_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.playlist_song_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.playlist_song_id_seq OWNER TO postgres;

--
-- Name: playlist_song_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.playlist_song_id_seq OWNED BY public.playlist_song.id;


--
-- Name: playlists; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.playlists (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    description character varying(191),
    user_id bigint NOT NULL,
    is_public boolean DEFAULT false NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    uuid uuid
);


ALTER TABLE public.playlists OWNER TO postgres;

--
-- Name: playlists_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.playlists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.playlists_id_seq OWNER TO postgres;

--
-- Name: playlists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.playlists_id_seq OWNED BY public.playlists.id;


--
-- Name: post_producer_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.post_producer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.post_producer_id_seq OWNER TO postgres;

--
-- Name: post_producer_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.post_producer_id_seq OWNED BY public.anime_producer.id;


--
-- Name: post_studio_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.post_studio_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.post_studio_id_seq OWNER TO postgres;

--
-- Name: post_studio_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.post_studio_id_seq OWNED BY public.anime_studio.id;


--
-- Name: posts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.posts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.posts_id_seq OWNER TO postgres;

--
-- Name: posts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.posts_id_seq OWNED BY public.animes.id;


--
-- Name: producers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.producers (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    logo character varying(191),
    uuid uuid,
    anime_count integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.producers OWNER TO postgres;

--
-- Name: producers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.producers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.producers_id_seq OWNER TO postgres;

--
-- Name: producers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.producers_id_seq OWNED BY public.producers.id;


--
-- Name: ranking_histories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ranking_histories (
    id bigint NOT NULL,
    song_id bigint NOT NULL,
    rank integer NOT NULL,
    seasonal_rank integer,
    score numeric(8,2),
    date date NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.ranking_histories OWNER TO postgres;

--
-- Name: ranking_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ranking_histories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ranking_histories_id_seq OWNER TO postgres;

--
-- Name: ranking_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ranking_histories_id_seq OWNED BY public.ranking_histories.id;


--
-- Name: song_ratings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.song_ratings (
    id bigint NOT NULL,
    rating integer NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    song_id bigint
);


ALTER TABLE public.song_ratings OWNER TO postgres;

--
-- Name: ratings_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ratings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ratings_id_seq OWNER TO postgres;

--
-- Name: ratings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ratings_id_seq OWNED BY public.song_ratings.id;


--
-- Name: song_reports; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.song_reports (
    id bigint NOT NULL,
    song_id bigint NOT NULL,
    user_id bigint NOT NULL,
    source character varying(191) NOT NULL,
    title character varying(191) NOT NULL,
    content text NOT NULL,
    status boolean DEFAULT false NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    CONSTRAINT reports_status_check CHECK (((status)::text = ANY (ARRAY[('fixed'::character varying)::text, ('pending'::character varying)::text])))
);


ALTER TABLE public.song_reports OWNER TO postgres;

--
-- Name: reports_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.reports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.reports_id_seq OWNER TO postgres;

--
-- Name: reports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.reports_id_seq OWNED BY public.song_reports.id;


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role_permissions (
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL
);


ALTER TABLE public.role_permissions OWNER TO postgres;

--
-- Name: role_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role_user (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    role_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.role_user OWNER TO postgres;

--
-- Name: role_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.role_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.role_user_id_seq OWNER TO postgres;

--
-- Name: role_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.role_user_id_seq OWNED BY public.role_user.id;


--
-- Name: roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.roles (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    description character varying(191),
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    weight integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.roles OWNER TO postgres;

--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.roles_id_seq OWNER TO postgres;

--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: score_formats; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.score_formats (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    description character varying(191),
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.score_formats OWNER TO postgres;

--
-- Name: score_formats_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.score_formats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.score_formats_id_seq OWNER TO postgres;

--
-- Name: score_formats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.score_formats_id_seq OWNED BY public.score_formats.id;


--
-- Name: seasons; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.seasons (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    current boolean DEFAULT false NOT NULL
);


ALTER TABLE public.seasons OWNER TO postgres;

--
-- Name: seasons_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.seasons_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.seasons_id_seq OWNER TO postgres;

--
-- Name: seasons_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.seasons_id_seq OWNED BY public.seasons.id;


--
-- Name: song_reactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.song_reactions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    song_id bigint NOT NULL,
    type smallint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.song_reactions OWNER TO postgres;

--
-- Name: song_reactions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.song_reactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.song_reactions_id_seq OWNER TO postgres;

--
-- Name: song_reactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.song_reactions_id_seq OWNED BY public.song_reactions.id;


--
-- Name: song_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.song_user (
    id bigint NOT NULL,
    song_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.song_user OWNER TO postgres;

--
-- Name: song_user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.song_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.song_user_id_seq OWNER TO postgres;

--
-- Name: song_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.song_user_id_seq OWNED BY public.song_user.id;


--
-- Name: song_variants; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.song_variants (
    id bigint NOT NULL,
    version_number bigint DEFAULT '1'::bigint NOT NULL,
    song_id bigint NOT NULL,
    views bigint DEFAULT '0'::bigint NOT NULL,
    slug character varying(191) NOT NULL,
    season_id bigint NOT NULL,
    year_id bigint NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    spoiler boolean DEFAULT false NOT NULL,
    status boolean DEFAULT false NOT NULL,
    anime_themes_id bigint
);


ALTER TABLE public.song_variants OWNER TO postgres;

--
-- Name: song_variants_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.song_variants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.song_variants_id_seq OWNER TO postgres;

--
-- Name: song_variants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.song_variants_id_seq OWNED BY public.song_variants.id;


--
-- Name: songs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.songs (
    id bigint NOT NULL,
    song_romaji character varying(191),
    song_jp character varying(191),
    song_en character varying(191),
    theme_num character varying(191) DEFAULT '1'::character varying NOT NULL,
    type character varying(255) DEFAULT 'OP'::character varying NOT NULL,
    slug character varying(191) NOT NULL,
    anime_id bigint NOT NULL,
    season_id bigint NOT NULL,
    year_id bigint NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    views bigint DEFAULT '0'::bigint NOT NULL,
    likes_count bigint DEFAULT '0'::bigint NOT NULL,
    dislikes_count bigint DEFAULT '0'::bigint NOT NULL,
    prev_main_rank integer,
    prev_seasonal_rank integer,
    status boolean DEFAULT false NOT NULL,
    favorites_count integer DEFAULT 0 NOT NULL,
    average_score numeric(5,2) DEFAULT '0'::numeric NOT NULL,
    uuid uuid,
    anime_themes_id bigint,
    CONSTRAINT songs_type_check CHECK (((type)::text = ANY ((ARRAY['OP'::character varying, 'ED'::character varying, 'INS'::character varying, 'OTH'::character varying])::text[])))
);


ALTER TABLE public.songs OWNER TO postgres;

--
-- Name: songs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.songs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.songs_id_seq OWNER TO postgres;

--
-- Name: songs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.songs_id_seq OWNED BY public.songs.id;


--
-- Name: studios; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.studios (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    logo character varying(191),
    uuid uuid,
    anime_count integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.studios OWNER TO postgres;

--
-- Name: studios_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.studios_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.studios_id_seq OWNER TO postgres;

--
-- Name: studios_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.studios_id_seq OWNED BY public.studios.id;


--
-- Name: tournament_matchups; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tournament_matchups (
    id bigint NOT NULL,
    tournament_id bigint NOT NULL,
    round integer NOT NULL,
    "position" integer NOT NULL,
    song1_id bigint,
    song2_id bigint,
    song1_votes integer DEFAULT 0 NOT NULL,
    song2_votes integer DEFAULT 0 NOT NULL,
    winner_song_id bigint,
    ends_at timestamp(0) without time zone,
    is_active boolean DEFAULT false NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.tournament_matchups OWNER TO postgres;

--
-- Name: tournament_matchups_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tournament_matchups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tournament_matchups_id_seq OWNER TO postgres;

--
-- Name: tournament_matchups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tournament_matchups_id_seq OWNED BY public.tournament_matchups.id;


--
-- Name: tournament_votes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tournament_votes (
    id bigint NOT NULL,
    tournament_matchup_id bigint NOT NULL,
    user_id bigint NOT NULL,
    song_id bigint NOT NULL,
    ip_address character varying(45),
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.tournament_votes OWNER TO postgres;

--
-- Name: tournament_votes_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tournament_votes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tournament_votes_id_seq OWNER TO postgres;

--
-- Name: tournament_votes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tournament_votes_id_seq OWNED BY public.tournament_votes.id;


--
-- Name: tournaments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tournaments (
    id bigint NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191) NOT NULL,
    description text,
    size integer DEFAULT 16 NOT NULL,
    status character varying(255) DEFAULT 'draft'::character varying NOT NULL,
    current_round integer,
    winner_song_id bigint,
    started_at timestamp(0) without time zone,
    completed_at timestamp(0) without time zone,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    type_filter character varying(191),
    uuid uuid,
    CONSTRAINT tournaments_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'active'::character varying, 'completed'::character varying])::text[])))
);


ALTER TABLE public.tournaments OWNER TO postgres;

--
-- Name: tournaments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tournaments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tournaments_id_seq OWNER TO postgres;

--
-- Name: tournaments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tournaments_id_seq OWNED BY public.tournaments.id;


--
-- Name: user_reports; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_reports (
    id bigint NOT NULL,
    reported_user_id bigint NOT NULL,
    reporter_user_id bigint NOT NULL,
    source character varying(50),
    reason character varying(191) NOT NULL,
    content text,
    status boolean DEFAULT false NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.user_reports OWNER TO postgres;

--
-- Name: user_reports_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_reports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_reports_id_seq OWNER TO postgres;

--
-- Name: user_reports_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_reports_id_seq OWNED BY public.user_reports.id;


--
-- Name: user_requests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_requests (
    id bigint NOT NULL,
    title character varying(191) NOT NULL,
    content text NOT NULL,
    user_id bigint NOT NULL,
    attended_by bigint,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    status boolean DEFAULT false NOT NULL
);


ALTER TABLE public.user_requests OWNER TO postgres;

--
-- Name: user_requests_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_requests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_requests_id_seq OWNER TO postgres;

--
-- Name: user_requests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_requests_id_seq OWNED BY public.user_requests.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    uuid uuid NOT NULL,
    name character varying(191) NOT NULL,
    slug character varying(191),
    email character varying(191) NOT NULL,
    email_verified_at timestamp(0) without time zone,
    password character varying(191) NOT NULL,
    last_login_at timestamp(0) without time zone,
    remember_token character varying(100),
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    avatar character varying(191),
    banner character varying(191),
    score_format_id bigint,
    xp bigint DEFAULT '0'::bigint NOT NULL,
    level integer DEFAULT 1 NOT NULL,
    anilist_id bigint,
    anilist_username character varying(191),
    anilist_access_token text,
    anilist_refresh_token text,
    anilist_token_expires_at timestamp(0) without time zone,
    google_id character varying(255),
    google_email character varying(255),
    google_access_token text,
    google_refresh_token text,
    google_token_expires_at timestamp(0) without time zone,
    profile_color character varying(191) DEFAULT '#7f13ec'::character varying,
    about text
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: videos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.videos (
    id bigint NOT NULL,
    embed_code text,
    video_src text,
    song_variant_id bigint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    status boolean DEFAULT false NOT NULL
);


ALTER TABLE public.videos OWNER TO postgres;

--
-- Name: videos_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.videos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.videos_id_seq OWNER TO postgres;

--
-- Name: videos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.videos_id_seq OWNED BY public.videos.id;


--
-- Name: xp_activities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.xp_activities (
    id bigint NOT NULL,
    key character varying(50) NOT NULL,
    xp_amount integer NOT NULL,
    description character varying(255),
    cooldown_seconds integer DEFAULT 0 NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.xp_activities OWNER TO postgres;

--
-- Name: xp_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.xp_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.xp_activities_id_seq OWNER TO postgres;

--
-- Name: xp_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.xp_activities_id_seq OWNED BY public.xp_activities.id;


--
-- Name: xp_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.xp_logs (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    xp_activity_id bigint NOT NULL,
    xp_amount integer NOT NULL,
    metadata json,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


ALTER TABLE public.xp_logs OWNER TO postgres;

--
-- Name: xp_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.xp_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.xp_logs_id_seq OWNER TO postgres;

--
-- Name: xp_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.xp_logs_id_seq OWNED BY public.xp_logs.id;


--
-- Name: years; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.years (
    id bigint NOT NULL,
    name smallint NOT NULL,
    slug character varying(191) NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    current boolean DEFAULT false NOT NULL
);


ALTER TABLE public.years OWNER TO postgres;

--
-- Name: years_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.years_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.years_id_seq OWNER TO postgres;

--
-- Name: years_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.years_id_seq OWNED BY public.years.id;


--
-- Name: activities id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.activities ALTER COLUMN id SET DEFAULT nextval('public.activities_id_seq'::regclass);


--
-- Name: anime_external_link id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_external_link ALTER COLUMN id SET DEFAULT nextval('public.external_link_post_id_seq'::regclass);


--
-- Name: anime_genre id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_genre ALTER COLUMN id SET DEFAULT nextval('public.genre_post_id_seq'::regclass);


--
-- Name: anime_producer id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_producer ALTER COLUMN id SET DEFAULT nextval('public.post_producer_id_seq'::regclass);


--
-- Name: anime_studio id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_studio ALTER COLUMN id SET DEFAULT nextval('public.post_studio_id_seq'::regclass);


--
-- Name: animes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes ALTER COLUMN id SET DEFAULT nextval('public.posts_id_seq'::regclass);


--
-- Name: announcements id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.announcements ALTER COLUMN id SET DEFAULT nextval('public.announcements_id_seq'::regclass);


--
-- Name: artist_song id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_song ALTER COLUMN id SET DEFAULT nextval('public.artist_song_id_seq'::regclass);


--
-- Name: artist_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_user ALTER COLUMN id SET DEFAULT nextval('public.artist_user_id_seq'::regclass);


--
-- Name: artists id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artists ALTER COLUMN id SET DEFAULT nextval('public.artists_id_seq'::regclass);


--
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);


--
-- Name: badge_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.badge_user ALTER COLUMN id SET DEFAULT nextval('public.badge_user_id_seq'::regclass);


--
-- Name: badges id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.badges ALTER COLUMN id SET DEFAULT nextval('public.badges_id_seq'::regclass);


--
-- Name: comment_reactions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reactions ALTER COLUMN id SET DEFAULT nextval('public.comment_reactions_id_seq'::regclass);


--
-- Name: comment_reports id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reports ALTER COLUMN id SET DEFAULT nextval('public.comment_reports_id_seq'::regclass);


--
-- Name: comments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments ALTER COLUMN id SET DEFAULT nextval('public.comments_id_seq'::regclass);


--
-- Name: daily_metrics id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_metrics ALTER COLUMN id SET DEFAULT nextval('public.daily_metrics_id_seq'::regclass);


--
-- Name: external_links id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.external_links ALTER COLUMN id SET DEFAULT nextval('public.external_links_id_seq'::regclass);


--
-- Name: failed_jobs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.failed_jobs ALTER COLUMN id SET DEFAULT nextval('public.failed_jobs_id_seq'::regclass);


--
-- Name: formats id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.formats ALTER COLUMN id SET DEFAULT nextval('public.formats_id_seq'::regclass);


--
-- Name: genres id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genres ALTER COLUMN id SET DEFAULT nextval('public.genres_id_seq'::regclass);


--
-- Name: migrations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migrations ALTER COLUMN id SET DEFAULT nextval('public.migrations_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: personal_access_tokens id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.personal_access_tokens ALTER COLUMN id SET DEFAULT nextval('public.personal_access_tokens_id_seq'::regclass);


--
-- Name: playlist_song id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist_song ALTER COLUMN id SET DEFAULT nextval('public.playlist_song_id_seq'::regclass);


--
-- Name: playlists id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlists ALTER COLUMN id SET DEFAULT nextval('public.playlists_id_seq'::regclass);


--
-- Name: producers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.producers ALTER COLUMN id SET DEFAULT nextval('public.producers_id_seq'::regclass);


--
-- Name: ranking_histories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ranking_histories ALTER COLUMN id SET DEFAULT nextval('public.ranking_histories_id_seq'::regclass);


--
-- Name: role_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_user ALTER COLUMN id SET DEFAULT nextval('public.role_user_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: score_formats id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.score_formats ALTER COLUMN id SET DEFAULT nextval('public.score_formats_id_seq'::regclass);


--
-- Name: seasons id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.seasons ALTER COLUMN id SET DEFAULT nextval('public.seasons_id_seq'::regclass);


--
-- Name: song_ratings id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_ratings ALTER COLUMN id SET DEFAULT nextval('public.ratings_id_seq'::regclass);


--
-- Name: song_reactions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reactions ALTER COLUMN id SET DEFAULT nextval('public.song_reactions_id_seq'::regclass);


--
-- Name: song_reports id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reports ALTER COLUMN id SET DEFAULT nextval('public.reports_id_seq'::regclass);


--
-- Name: song_user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_user ALTER COLUMN id SET DEFAULT nextval('public.song_user_id_seq'::regclass);


--
-- Name: song_variants id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_variants ALTER COLUMN id SET DEFAULT nextval('public.song_variants_id_seq'::regclass);


--
-- Name: songs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.songs ALTER COLUMN id SET DEFAULT nextval('public.songs_id_seq'::regclass);


--
-- Name: studios id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.studios ALTER COLUMN id SET DEFAULT nextval('public.studios_id_seq'::regclass);


--
-- Name: tournament_matchups id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_matchups ALTER COLUMN id SET DEFAULT nextval('public.tournament_matchups_id_seq'::regclass);


--
-- Name: tournament_votes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_votes ALTER COLUMN id SET DEFAULT nextval('public.tournament_votes_id_seq'::regclass);


--
-- Name: tournaments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournaments ALTER COLUMN id SET DEFAULT nextval('public.tournaments_id_seq'::regclass);


--
-- Name: user_reports id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_reports ALTER COLUMN id SET DEFAULT nextval('public.user_reports_id_seq'::regclass);


--
-- Name: user_requests id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_requests ALTER COLUMN id SET DEFAULT nextval('public.user_requests_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: videos id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.videos ALTER COLUMN id SET DEFAULT nextval('public.videos_id_seq'::regclass);


--
-- Name: xp_activities id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xp_activities ALTER COLUMN id SET DEFAULT nextval('public.xp_activities_id_seq'::regclass);


--
-- Name: xp_logs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xp_logs ALTER COLUMN id SET DEFAULT nextval('public.xp_logs_id_seq'::regclass);


--
-- Name: years id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.years ALTER COLUMN id SET DEFAULT nextval('public.years_id_seq'::regclass);


--
-- Name: activities activities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_pkey PRIMARY KEY (id);


--
-- Name: animes animes_anilist_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes
    ADD CONSTRAINT animes_anilist_id_unique UNIQUE (anilist_id);


--
-- Name: animes animes_anime_themes_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes
    ADD CONSTRAINT animes_anime_themes_id_unique UNIQUE (anime_themes_id);


--
-- Name: animes animes_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes
    ADD CONSTRAINT animes_uuid_unique UNIQUE (uuid);


--
-- Name: announcements announcements_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.announcements
    ADD CONSTRAINT announcements_pkey PRIMARY KEY (id);


--
-- Name: artist_song artist_song_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_song
    ADD CONSTRAINT artist_song_pkey PRIMARY KEY (id);


--
-- Name: artist_user artist_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_user
    ADD CONSTRAINT artist_user_pkey PRIMARY KEY (id);


--
-- Name: artists artists_anilist_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artists
    ADD CONSTRAINT artists_anilist_id_unique UNIQUE (anilist_id);


--
-- Name: artists artists_anime_themes_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artists
    ADD CONSTRAINT artists_anime_themes_id_unique UNIQUE (anime_themes_id);


--
-- Name: artists artists_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artists
    ADD CONSTRAINT artists_pkey PRIMARY KEY (id);


--
-- Name: artists artists_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artists
    ADD CONSTRAINT artists_slug_unique UNIQUE (slug);


--
-- Name: artists artists_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artists
    ADD CONSTRAINT artists_uuid_unique UNIQUE (uuid);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: badge_user badge_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.badge_user
    ADD CONSTRAINT badge_user_pkey PRIMARY KEY (id);


--
-- Name: badges badges_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.badges
    ADD CONSTRAINT badges_pkey PRIMARY KEY (id);


--
-- Name: cache_locks cache_locks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cache_locks
    ADD CONSTRAINT cache_locks_pkey PRIMARY KEY (key);


--
-- Name: cache cache_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cache
    ADD CONSTRAINT cache_pkey PRIMARY KEY (key);


--
-- Name: comment_reactions comment_reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reactions
    ADD CONSTRAINT comment_reactions_pkey PRIMARY KEY (id);


--
-- Name: comment_reactions comment_reactions_user_id_comment_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reactions
    ADD CONSTRAINT comment_reactions_user_id_comment_id_unique UNIQUE (user_id, comment_id);


--
-- Name: comment_reports comment_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reports
    ADD CONSTRAINT comment_reports_pkey PRIMARY KEY (id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: daily_metrics daily_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_metrics
    ADD CONSTRAINT daily_metrics_pkey PRIMARY KEY (id);


--
-- Name: daily_metrics daily_metrics_song_id_date_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_metrics
    ADD CONSTRAINT daily_metrics_song_id_date_unique UNIQUE (song_id, date);


--
-- Name: anime_external_link external_link_post_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_external_link
    ADD CONSTRAINT external_link_post_pkey PRIMARY KEY (id);


--
-- Name: external_links external_links_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.external_links
    ADD CONSTRAINT external_links_pkey PRIMARY KEY (id);


--
-- Name: failed_jobs failed_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.failed_jobs
    ADD CONSTRAINT failed_jobs_pkey PRIMARY KEY (id);


--
-- Name: failed_jobs failed_jobs_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.failed_jobs
    ADD CONSTRAINT failed_jobs_uuid_unique UNIQUE (uuid);


--
-- Name: follows follows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_pkey PRIMARY KEY (follower_id, followed_id);


--
-- Name: formats formats_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.formats
    ADD CONSTRAINT formats_pkey PRIMARY KEY (id);


--
-- Name: anime_genre genre_post_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_genre
    ADD CONSTRAINT genre_post_pkey PRIMARY KEY (id);


--
-- Name: genres genres_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_pkey PRIMARY KEY (id);


--
-- Name: genres genres_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_slug_unique UNIQUE (slug);


--
-- Name: levels levels_min_xp_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.levels
    ADD CONSTRAINT levels_min_xp_unique UNIQUE (min_xp);


--
-- Name: levels levels_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.levels
    ADD CONSTRAINT levels_pkey PRIMARY KEY (level);


--
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_slug_unique UNIQUE (slug);


--
-- Name: personal_access_tokens personal_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.personal_access_tokens
    ADD CONSTRAINT personal_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: personal_access_tokens personal_access_tokens_token_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.personal_access_tokens
    ADD CONSTRAINT personal_access_tokens_token_unique UNIQUE (token);


--
-- Name: playlist_song playlist_song_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist_song
    ADD CONSTRAINT playlist_song_pkey PRIMARY KEY (id);


--
-- Name: playlist_song playlist_song_playlist_id_song_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist_song
    ADD CONSTRAINT playlist_song_playlist_id_song_id_unique UNIQUE (playlist_id, song_id);


--
-- Name: playlists playlists_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlists
    ADD CONSTRAINT playlists_pkey PRIMARY KEY (id);


--
-- Name: playlists playlists_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlists
    ADD CONSTRAINT playlists_uuid_unique UNIQUE (uuid);


--
-- Name: anime_producer post_producer_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_producer
    ADD CONSTRAINT post_producer_pkey PRIMARY KEY (id);


--
-- Name: anime_studio post_studio_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_studio
    ADD CONSTRAINT post_studio_pkey PRIMARY KEY (id);


--
-- Name: animes posts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);


--
-- Name: animes posts_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes
    ADD CONSTRAINT posts_slug_unique UNIQUE (slug);


--
-- Name: producers producers_name_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.producers
    ADD CONSTRAINT producers_name_unique UNIQUE (name);


--
-- Name: producers producers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.producers
    ADD CONSTRAINT producers_pkey PRIMARY KEY (id);


--
-- Name: producers producers_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.producers
    ADD CONSTRAINT producers_slug_unique UNIQUE (slug);


--
-- Name: producers producers_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.producers
    ADD CONSTRAINT producers_uuid_unique UNIQUE (uuid);


--
-- Name: ranking_histories ranking_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ranking_histories
    ADD CONSTRAINT ranking_histories_pkey PRIMARY KEY (id);


--
-- Name: ranking_histories ranking_histories_song_id_date_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ranking_histories
    ADD CONSTRAINT ranking_histories_song_id_date_unique UNIQUE (song_id, date);


--
-- Name: song_ratings ratings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_ratings
    ADD CONSTRAINT ratings_pkey PRIMARY KEY (id);


--
-- Name: song_ratings ratings_user_song_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_ratings
    ADD CONSTRAINT ratings_user_song_unique UNIQUE (user_id, song_id);


--
-- Name: song_reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: role_user role_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_user
    ADD CONSTRAINT role_user_pkey PRIMARY KEY (id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: roles roles_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_slug_unique UNIQUE (slug);


--
-- Name: score_formats score_formats_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.score_formats
    ADD CONSTRAINT score_formats_pkey PRIMARY KEY (id);


--
-- Name: score_formats score_formats_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.score_formats
    ADD CONSTRAINT score_formats_slug_unique UNIQUE (slug);


--
-- Name: seasons seasons_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.seasons
    ADD CONSTRAINT seasons_pkey PRIMARY KEY (id);


--
-- Name: seasons seasons_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.seasons
    ADD CONSTRAINT seasons_slug_unique UNIQUE (slug);


--
-- Name: song_reactions song_reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reactions
    ADD CONSTRAINT song_reactions_pkey PRIMARY KEY (id);


--
-- Name: song_reactions song_reactions_user_id_song_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reactions
    ADD CONSTRAINT song_reactions_user_id_song_id_unique UNIQUE (user_id, song_id);


--
-- Name: song_user song_user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_user
    ADD CONSTRAINT song_user_pkey PRIMARY KEY (id);


--
-- Name: song_variants song_variants_anime_themes_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_variants
    ADD CONSTRAINT song_variants_anime_themes_id_unique UNIQUE (anime_themes_id);


--
-- Name: song_variants song_variants_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_variants
    ADD CONSTRAINT song_variants_pkey PRIMARY KEY (id);


--
-- Name: songs songs_anime_themes_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.songs
    ADD CONSTRAINT songs_anime_themes_id_unique UNIQUE (anime_themes_id);


--
-- Name: songs songs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.songs
    ADD CONSTRAINT songs_pkey PRIMARY KEY (id);


--
-- Name: songs songs_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.songs
    ADD CONSTRAINT songs_uuid_unique UNIQUE (uuid);


--
-- Name: studios studios_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.studios
    ADD CONSTRAINT studios_pkey PRIMARY KEY (id);


--
-- Name: studios studios_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.studios
    ADD CONSTRAINT studios_uuid_unique UNIQUE (uuid);


--
-- Name: tournament_matchups tournament_matchups_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_matchups
    ADD CONSTRAINT tournament_matchups_pkey PRIMARY KEY (id);


--
-- Name: tournament_votes tournament_votes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_votes
    ADD CONSTRAINT tournament_votes_pkey PRIMARY KEY (id);


--
-- Name: tournaments tournaments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournaments
    ADD CONSTRAINT tournaments_pkey PRIMARY KEY (id);


--
-- Name: tournaments tournaments_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournaments
    ADD CONSTRAINT tournaments_slug_unique UNIQUE (slug);


--
-- Name: tournaments tournaments_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournaments
    ADD CONSTRAINT tournaments_uuid_unique UNIQUE (uuid);


--
-- Name: tournament_votes unique_user_matchup_vote; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_votes
    ADD CONSTRAINT unique_user_matchup_vote UNIQUE (tournament_matchup_id, user_id);


--
-- Name: user_reports user_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_reports
    ADD CONSTRAINT user_reports_pkey PRIMARY KEY (id);


--
-- Name: user_requests user_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_requests
    ADD CONSTRAINT user_requests_pkey PRIMARY KEY (id);


--
-- Name: users users_anilist_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_anilist_id_unique UNIQUE (anilist_id);


--
-- Name: users users_email_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_unique UNIQUE (email);


--
-- Name: users users_google_email_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_google_email_unique UNIQUE (google_email);


--
-- Name: users users_google_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_google_id_unique UNIQUE (google_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_uuid_unique UNIQUE (uuid);


--
-- Name: videos videos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.videos
    ADD CONSTRAINT videos_pkey PRIMARY KEY (id);


--
-- Name: xp_activities xp_activities_key_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xp_activities
    ADD CONSTRAINT xp_activities_key_unique UNIQUE (key);


--
-- Name: xp_activities xp_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xp_activities
    ADD CONSTRAINT xp_activities_pkey PRIMARY KEY (id);


--
-- Name: xp_logs xp_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xp_logs
    ADD CONSTRAINT xp_logs_pkey PRIMARY KEY (id);


--
-- Name: years years_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.years
    ADD CONSTRAINT years_pkey PRIMARY KEY (id);


--
-- Name: years years_slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.years
    ADD CONSTRAINT years_slug_unique UNIQUE (slug);


--
-- Name: activities_user_id_created_at_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX activities_user_id_created_at_index ON public.activities USING btree (user_id, created_at);


--
-- Name: animes_title_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX animes_title_trgm_idx ON public.animes USING gin (title public.gin_trgm_ops);


--
-- Name: artists_name_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX artists_name_trgm_idx ON public.artists USING gin (name public.gin_trgm_ops);


--
-- Name: comment_reactions_type_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX comment_reactions_type_index ON public.comment_reactions USING btree (type);


--
-- Name: daily_metrics_site_wide_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX daily_metrics_site_wide_unique ON public.daily_metrics USING btree (date) WHERE (song_id IS NULL);


--
-- Name: genres_name_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX genres_name_index ON public.genres USING btree (name);


--
-- Name: idx_audit_event_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_audit_event_date ON public.audit_logs USING btree (event, created_at);


--
-- Name: idx_audit_polymorphic; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_audit_polymorphic ON public.audit_logs USING btree (auditable_type, auditable_id);


--
-- Name: idx_audit_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_audit_user ON public.audit_logs USING btree (user_id);


--
-- Name: idx_followed_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_followed_user ON public.follows USING btree (followed_id);


--
-- Name: idx_notif_subject; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_notif_subject ON public.notifications USING btree (subject_type, subject_id);


--
-- Name: idx_notif_user_unread; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_notif_user_unread ON public.notifications USING btree (user_id, read_at, created_at);


--
-- Name: idx_songs_ranks; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_songs_ranks ON public.songs USING btree (prev_main_rank, prev_seasonal_rank);


--
-- Name: idx_users_google_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_google_email ON public.users USING btree (google_email);


--
-- Name: idx_users_google_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_google_id ON public.users USING btree (google_id);


--
-- Name: password_resets_email_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX password_resets_email_index ON public.password_resets USING btree (email);


--
-- Name: personal_access_tokens_tokenable_type_tokenable_id_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX personal_access_tokens_tokenable_type_tokenable_id_index ON public.personal_access_tokens USING btree (tokenable_type, tokenable_id);


--
-- Name: posts_status_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX posts_status_idx ON public.animes USING btree (status);


--
-- Name: seasons_current_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX seasons_current_index ON public.seasons USING btree (current);


--
-- Name: song_ratings_created_at_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX song_ratings_created_at_index ON public.song_ratings USING btree (created_at);


--
-- Name: song_reactions_type_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX song_reactions_type_index ON public.song_reactions USING btree (type);


--
-- Name: songs_created_at_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX songs_created_at_index ON public.songs USING btree (created_at);


--
-- Name: songs_en_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX songs_en_trgm_idx ON public.songs USING gin (song_en public.gin_trgm_ops);


--
-- Name: songs_jp_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX songs_jp_trgm_idx ON public.songs USING gin (song_jp public.gin_trgm_ops);


--
-- Name: songs_romaji_trgm_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX songs_romaji_trgm_idx ON public.songs USING gin (song_romaji public.gin_trgm_ops);


--
-- Name: songs_slug_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX songs_slug_idx ON public.songs USING btree (slug);


--
-- Name: songs_type_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX songs_type_idx ON public.songs USING btree (type);


--
-- Name: songs_views_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX songs_views_idx ON public.songs USING btree (views);


--
-- Name: user_reports_reported_user_id_status_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_reports_reported_user_id_status_index ON public.user_reports USING btree (reported_user_id, status);


--
-- Name: user_reports_reporter_user_id_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_reports_reporter_user_id_index ON public.user_reports USING btree (reporter_user_id);


--
-- Name: users_created_at_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX users_created_at_index ON public.users USING btree (created_at);


--
-- Name: years_current_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX years_current_index ON public.years USING btree (current);


--
-- Name: years_name_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX years_name_index ON public.years USING btree (name);


--
-- Name: artist_song trg_recount_artist_on_pivot_change; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_recount_artist_on_pivot_change AFTER INSERT OR DELETE OR UPDATE ON public.artist_song FOR EACH ROW EXECUTE FUNCTION public.handle_artist_song_change();


--
-- Name: songs trg_recount_artists_on_status_change; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_recount_artists_on_status_change AFTER UPDATE OF status ON public.songs FOR EACH ROW EXECUTE FUNCTION public.handle_song_status_change();


--
-- Name: songs trg_update_anime_song_counts; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_update_anime_song_counts AFTER INSERT OR DELETE OR UPDATE ON public.songs FOR EACH ROW EXECUTE FUNCTION public.update_anime_song_counts();


--
-- Name: songs trg_update_anime_songs_count; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_update_anime_songs_count AFTER INSERT OR DELETE OR UPDATE ON public.songs FOR EACH ROW EXECUTE FUNCTION public.update_anime_songs_count();


--
-- Name: artist_user trg_update_artist_favorites_count; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_update_artist_favorites_count AFTER INSERT OR DELETE ON public.artist_user FOR EACH ROW EXECUTE FUNCTION public.update_artist_favorites_count();


--
-- Name: song_ratings trg_update_song_average_score; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_update_song_average_score AFTER INSERT OR DELETE OR UPDATE ON public.song_ratings FOR EACH ROW EXECUTE FUNCTION public.update_song_average_score();


--
-- Name: song_user trg_update_song_favorites_count; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_update_song_favorites_count AFTER INSERT OR DELETE ON public.song_user FOR EACH ROW EXECUTE FUNCTION public.update_song_favorites_count();


--
-- Name: songs trig_song_views_update; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trig_song_views_update AFTER UPDATE OF views ON public.songs FOR EACH ROW WHEN ((new.views > old.views)) EXECUTE FUNCTION public.fn_update_daily_song_views();


--
-- Name: song_variants trig_variant_views_update; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trig_variant_views_update AFTER UPDATE OF views ON public.song_variants FOR EACH ROW WHEN ((new.views > old.views)) EXECUTE FUNCTION public.fn_update_daily_variant_views();


--
-- Name: activities activities_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: anime_external_link anime_external_link_anime_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_external_link
    ADD CONSTRAINT anime_external_link_anime_id_foreign FOREIGN KEY (anime_id) REFERENCES public.animes(id) ON DELETE CASCADE;


--
-- Name: anime_genre anime_genre_anime_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_genre
    ADD CONSTRAINT anime_genre_anime_id_foreign FOREIGN KEY (anime_id) REFERENCES public.animes(id) ON DELETE CASCADE;


--
-- Name: anime_producer anime_producer_anime_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_producer
    ADD CONSTRAINT anime_producer_anime_id_foreign FOREIGN KEY (anime_id) REFERENCES public.animes(id) ON DELETE CASCADE;


--
-- Name: anime_studio anime_studio_anime_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_studio
    ADD CONSTRAINT anime_studio_anime_id_foreign FOREIGN KEY (anime_id) REFERENCES public.animes(id) ON DELETE CASCADE;


--
-- Name: artist_song artist_song_artist_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_song
    ADD CONSTRAINT artist_song_artist_id_foreign FOREIGN KEY (artist_id) REFERENCES public.artists(id) ON DELETE CASCADE;


--
-- Name: artist_song artist_song_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_song
    ADD CONSTRAINT artist_song_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: artist_user artist_user_artist_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_user
    ADD CONSTRAINT artist_user_artist_id_foreign FOREIGN KEY (artist_id) REFERENCES public.artists(id) ON DELETE CASCADE;


--
-- Name: artist_user artist_user_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_user
    ADD CONSTRAINT artist_user_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: audit_logs audit_logs_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: badge_user badge_user_badge_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.badge_user
    ADD CONSTRAINT badge_user_badge_id_foreign FOREIGN KEY (badge_id) REFERENCES public.badges(id) ON DELETE CASCADE;


--
-- Name: badge_user badge_user_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.badge_user
    ADD CONSTRAINT badge_user_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: comment_reactions comment_reactions_comment_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reactions
    ADD CONSTRAINT comment_reactions_comment_id_foreign FOREIGN KEY (comment_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: comment_reactions comment_reactions_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reactions
    ADD CONSTRAINT comment_reactions_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: comment_reports comment_reports_comment_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reports
    ADD CONSTRAINT comment_reports_comment_id_foreign FOREIGN KEY (comment_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: comment_reports comment_reports_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comment_reports
    ADD CONSTRAINT comment_reports_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: comments comments_parent_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_parent_id_foreign FOREIGN KEY (parent_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: comments comments_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: comments comments_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: daily_metrics daily_metrics_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_metrics
    ADD CONSTRAINT daily_metrics_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: anime_external_link external_link_post_external_link_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_external_link
    ADD CONSTRAINT external_link_post_external_link_id_foreign FOREIGN KEY (external_link_id) REFERENCES public.external_links(id) ON DELETE CASCADE;


--
-- Name: follows follows_followed_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_followed_id_foreign FOREIGN KEY (followed_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: follows follows_follower_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_follower_id_foreign FOREIGN KEY (follower_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: anime_genre genre_post_genre_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_genre
    ADD CONSTRAINT genre_post_genre_id_foreign FOREIGN KEY (genre_id) REFERENCES public.genres(id) ON DELETE CASCADE;


--
-- Name: levels levels_badge_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.levels
    ADD CONSTRAINT levels_badge_id_foreign FOREIGN KEY (badge_id) REFERENCES public.badges(id) ON DELETE SET NULL;


--
-- Name: notifications notifications_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: playlist_song playlist_song_playlist_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist_song
    ADD CONSTRAINT playlist_song_playlist_id_foreign FOREIGN KEY (playlist_id) REFERENCES public.playlists(id) ON DELETE CASCADE;


--
-- Name: playlist_song playlist_song_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlist_song
    ADD CONSTRAINT playlist_song_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: playlists playlists_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.playlists
    ADD CONSTRAINT playlists_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: anime_producer post_producer_producer_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_producer
    ADD CONSTRAINT post_producer_producer_id_foreign FOREIGN KEY (producer_id) REFERENCES public.producers(id) ON DELETE CASCADE;


--
-- Name: anime_studio post_studio_studio_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.anime_studio
    ADD CONSTRAINT post_studio_studio_id_foreign FOREIGN KEY (studio_id) REFERENCES public.studios(id) ON DELETE CASCADE;


--
-- Name: animes posts_format_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes
    ADD CONSTRAINT posts_format_id_foreign FOREIGN KEY (format_id) REFERENCES public.formats(id) ON DELETE SET NULL;


--
-- Name: animes posts_season_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes
    ADD CONSTRAINT posts_season_id_foreign FOREIGN KEY (season_id) REFERENCES public.seasons(id) ON DELETE SET NULL;


--
-- Name: animes posts_year_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.animes
    ADD CONSTRAINT posts_year_id_foreign FOREIGN KEY (year_id) REFERENCES public.years(id) ON DELETE SET NULL;


--
-- Name: ranking_histories ranking_histories_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ranking_histories
    ADD CONSTRAINT ranking_histories_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: song_ratings ratings_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_ratings
    ADD CONSTRAINT ratings_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: song_reports reports_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reports
    ADD CONSTRAINT reports_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: song_reports reports_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reports
    ADD CONSTRAINT reports_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_permission_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_foreign FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_foreign FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: role_user role_user_role_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_user
    ADD CONSTRAINT role_user_role_id_foreign FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: role_user role_user_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_user
    ADD CONSTRAINT role_user_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: song_ratings song_ratings_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_ratings
    ADD CONSTRAINT song_ratings_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: song_reactions song_reactions_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reactions
    ADD CONSTRAINT song_reactions_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: song_reactions song_reactions_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_reactions
    ADD CONSTRAINT song_reactions_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: song_user song_user_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_user
    ADD CONSTRAINT song_user_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: song_user song_user_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_user
    ADD CONSTRAINT song_user_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: song_variants song_variants_season_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_variants
    ADD CONSTRAINT song_variants_season_id_foreign FOREIGN KEY (season_id) REFERENCES public.seasons(id) ON DELETE CASCADE;


--
-- Name: song_variants song_variants_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_variants
    ADD CONSTRAINT song_variants_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: song_variants song_variants_year_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.song_variants
    ADD CONSTRAINT song_variants_year_id_foreign FOREIGN KEY (year_id) REFERENCES public.years(id) ON DELETE CASCADE;


--
-- Name: songs songs_anime_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.songs
    ADD CONSTRAINT songs_anime_id_foreign FOREIGN KEY (anime_id) REFERENCES public.animes(id) ON DELETE CASCADE;


--
-- Name: songs songs_season_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.songs
    ADD CONSTRAINT songs_season_id_foreign FOREIGN KEY (season_id) REFERENCES public.seasons(id) ON DELETE CASCADE;


--
-- Name: songs songs_year_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.songs
    ADD CONSTRAINT songs_year_id_foreign FOREIGN KEY (year_id) REFERENCES public.years(id) ON DELETE CASCADE;


--
-- Name: tournament_matchups tournament_matchups_song1_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_matchups
    ADD CONSTRAINT tournament_matchups_song1_id_foreign FOREIGN KEY (song1_id) REFERENCES public.songs(id) ON DELETE SET NULL;


--
-- Name: tournament_matchups tournament_matchups_song2_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_matchups
    ADD CONSTRAINT tournament_matchups_song2_id_foreign FOREIGN KEY (song2_id) REFERENCES public.songs(id) ON DELETE SET NULL;


--
-- Name: tournament_matchups tournament_matchups_tournament_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_matchups
    ADD CONSTRAINT tournament_matchups_tournament_id_foreign FOREIGN KEY (tournament_id) REFERENCES public.tournaments(id) ON DELETE CASCADE;


--
-- Name: tournament_matchups tournament_matchups_winner_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_matchups
    ADD CONSTRAINT tournament_matchups_winner_song_id_foreign FOREIGN KEY (winner_song_id) REFERENCES public.songs(id) ON DELETE SET NULL;


--
-- Name: tournament_votes tournament_votes_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_votes
    ADD CONSTRAINT tournament_votes_song_id_foreign FOREIGN KEY (song_id) REFERENCES public.songs(id) ON DELETE CASCADE;


--
-- Name: tournament_votes tournament_votes_tournament_matchup_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_votes
    ADD CONSTRAINT tournament_votes_tournament_matchup_id_foreign FOREIGN KEY (tournament_matchup_id) REFERENCES public.tournament_matchups(id) ON DELETE CASCADE;


--
-- Name: tournament_votes tournament_votes_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_votes
    ADD CONSTRAINT tournament_votes_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: tournaments tournaments_winner_song_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournaments
    ADD CONSTRAINT tournaments_winner_song_id_foreign FOREIGN KEY (winner_song_id) REFERENCES public.songs(id) ON DELETE SET NULL;


--
-- Name: user_reports user_reports_reported_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_reports
    ADD CONSTRAINT user_reports_reported_user_id_foreign FOREIGN KEY (reported_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_reports user_reports_reporter_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_reports
    ADD CONSTRAINT user_reports_reporter_user_id_foreign FOREIGN KEY (reporter_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_requests user_requests_attended_by_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_requests
    ADD CONSTRAINT user_requests_attended_by_foreign FOREIGN KEY (attended_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_requests user_requests_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_requests
    ADD CONSTRAINT user_requests_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users users_score_format_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_score_format_id_foreign FOREIGN KEY (score_format_id) REFERENCES public.score_formats(id) ON DELETE RESTRICT;


--
-- Name: videos videos_song_variant_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.videos
    ADD CONSTRAINT videos_song_variant_id_foreign FOREIGN KEY (song_variant_id) REFERENCES public.song_variants(id) ON DELETE CASCADE;


--
-- Name: xp_logs xp_logs_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xp_logs
    ADD CONSTRAINT xp_logs_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: xp_logs xp_logs_xp_activity_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xp_logs
    ADD CONSTRAINT xp_logs_xp_activity_id_foreign FOREIGN KEY (xp_activity_id) REFERENCES public.xp_activities(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict sLw193cihpID2OBD9TtbiMEBIdmc8hwgcSXeeE144gmcSkkufWdKYRJRKoEwRju

