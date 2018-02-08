port module Graph
    exposing
        ( draw
        , Model
        , new
        , Config
        , Msg(Draw)
        , subscriptions
        , update
        , view
        )

import Api
import Html exposing (Html)
import Html.Attributes as Attr
import Window
import Task
import Process
import Http
import Time exposing (Time)


-- CONFIG


type alias Config model msg =
    { get : model -> Model
    , set : model -> Model -> model
    , wrap : Msg -> msg
    }



-- MODEL


type alias Info =
    -- Info keeps track of state required to redraw a graph on changes of width
    { metric : String
    , width : Int
    }


type State
    = Delay String -- no size yet
    | Requesting Info -- requesting data from api
    | Drawing Info -- drawing the data
    | Finished Info (Result String DrawResult) -- done


type Model
    = Model
        { state : Maybe State -- drawing state
        , size : Maybe Window.Size -- size of the window
        , counter : Int -- how many resizes
        }


new : Config model msg -> ( Model, Cmd msg )
new config =
    ( Model
        { state = Nothing
        , size = Nothing
        , counter = 0
        }
    , Cmd.map config.wrap <| Task.perform Resize Window.size
    )



-- MSG


type Msg
    = Draw String
    | RenderResponse (Result Http.Error String)
    | Done DrawResult
    | Resize Window.Size
    | ResizeTimer Int



-- PORTS


type alias DrawRequest =
    { data : String
    }


port draw : DrawRequest -> Cmd msg


type alias DrawResult =
    { ok : Bool
    , error : String
    }


port done : (DrawResult -> msg) -> Sub msg



-- SUBSCRIPTIONS


subscriptions : Config model msg -> model -> Sub msg
subscriptions config model =
    doSubscriptions (config.get model)
        |> Sub.map config.wrap


doSubscriptions : Model -> Sub Msg
doSubscriptions (Model model) =
    Sub.batch
        [ done Done
        , Window.resizes Resize
        ]



-- UPDATE


update : Config model msg -> model -> Msg -> ( model, Cmd msg )
update config model msg =
    doUpdate config (config.get model) msg
        |> Tuple.mapFirst (config.set model)
        |> Tuple.mapSecond (Cmd.map config.wrap)


delay : Time
delay =
    (250 * Time.millisecond)


doUpdate : Config model msg -> Model -> Msg -> ( Model, Cmd Msg )
doUpdate config (Model model) msg =
    case msg of
        Draw metric ->
            let
                info width =
                    { metric = metric, width = width }

                do width =
                    ( Model { model | state = Just <| Requesting (info width) }
                    , Api.renderRequest metric
                        |> (\request -> { request | width = Just (width - 30) })
                        |> Api.render
                        |> Http.send RenderResponse
                    )
            in
                case ( model.size, model.state ) of
                    ( Just { width }, Nothing ) ->
                        do width

                    ( Just { width }, Just (Finished _ _) ) ->
                        do width

                    ( Just { width }, Just (Delay _) ) ->
                        do width

                    ( Nothing, _ ) ->
                        ( Model { model | state = Just <| Delay metric }
                        , Cmd.none
                        )

                    _ ->
                        ( Model model
                        , Cmd.none
                        )

        RenderResponse response ->
            case ( response, model.state ) of
                ( Ok data, Just (Requesting info) ) ->
                    ( Model { model | state = Just <| Drawing info }
                    , draw { data = data }
                    )

                ( Err error, Just (Requesting info) ) ->
                    let
                        state =
                            Just <| Finished info (Err <| toString error)
                    in
                        ( Model { model | state = state }
                        , Cmd.none
                        )

                _ ->
                    ( Model { model | state = Nothing }
                    , Cmd.none
                    )

        Done result ->
            case model.state of
                Just (Drawing info) ->
                    let
                        state =
                            Just <| Finished info (Ok result)
                    in
                        ( Model { model | state = state }
                        , Cmd.none
                        )

                _ ->
                    ( Model { model | state = Nothing }
                    , Cmd.none
                    )

        Resize size ->
            let
                newCounter =
                    model.counter + 1
            in
                ( Model { model | size = Just size, counter = newCounter }
                , case ( model.size, model.state ) of
                    ( _, Just (Delay metric) ) ->
                        sendMessage (Draw metric)

                    ( Just { width }, Just (Finished _ _) ) ->
                        if width /= size.width then
                            sendMessageAfter delay (ResizeTimer newCounter)
                        else
                            Cmd.none

                    _ ->
                        Cmd.none
                )

        ResizeTimer counter ->
            ( Model model
            , case ( counter == model.counter, model.state, model.size ) of
                ( True, Just (Finished info _), Just { width } ) ->
                    if info.width /= width then
                        sendMessage (Draw info.metric)
                    else
                        Cmd.none

                _ ->
                    Cmd.none
            )



-- VIEW


view : Config model msg -> Model -> Html msg
view config model =
    doView model
        |> Html.map config.wrap


doView : Model -> Html Msg
doView (Model model) =
    Html.div [ Attr.id "canvas-container" ]
        [ Html.canvas [ Attr.id "canvas" ] [] ]



-- UTILS


sendMessage : msg -> Cmd msg
sendMessage msg =
    Task.perform (always msg) (Task.succeed ())


sendMessageAfter : Time -> msg -> Cmd msg
sendMessageAfter delay msg =
    Task.perform (always msg) (Process.sleep delay)
